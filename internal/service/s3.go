package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aveplen-bach/s3/internal/model"
	"github.com/minio/minio-go/v7"
)

type minioImplementation struct {
	minioClient *minio.Client
	bucketName  string

	imagePathPrefix string
}

// GetImageObject implements S3ImageObjectService
func (i minioImplementation) GetImageObject(ctx context.Context, id uint64) (model.ImageObject, error) {
	var image model.ImageObject

	filename := i.getImageObjectName(id)

	_, err := i.minioClient.StatObject(
		ctx,
		i.bucketName,
		filename,
		minio.StatObjectOptions{},
	)
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "AccessDenied" {
			return image, fmt.Errorf(
				"access denied: %w",
				err,
			)
		}
		if errResponse.Code == "NoSuchBucket" {
			return image, fmt.Errorf(
				"bucket %s does not exist: %w",
				i.bucketName,
				err,
			)
		}
		if errResponse.Code == "InvalidBucketName" {
			return image, fmt.Errorf(
				"invalid bucket name – %s: %w",
				i.bucketName,
				err,
			)
		}
		if errResponse.Code == "NoSuchKey" {
			return image, fmt.Errorf(
				"object %s was not found: %w",
				filename,
				err,
			)
		}
	}

	object, err := i.minioClient.GetObject(
		ctx,
		i.bucketName,
		i.getImageObjectName(id),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return image, fmt.Errorf(
			"failed to get s3 object %s: %w",
			filename,
			err,
		)
	}

	contents, err := io.ReadAll(object)
	if err != nil {
		return image, fmt.Errorf(
			"failed to read object: %w",
			err,
		)
	}

	return model.ImageObject{
		Id:       id,
		Contents: contents,
	}, nil
}

// PutImageObject implements S3ImageObjectService
func (i minioImplementation) PutImageObject(ctx context.Context, object model.ImageObject) error {
	filename := i.getImageObjectName(object.Id)

	if _, err := i.minioClient.PutObject(
		ctx,
		i.bucketName,
		filename,
		bytes.NewReader(object.Contents),
		int64(len(object.Contents)),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	); err != nil {
		return fmt.Errorf(
			"failed to upload s3 object %s: %w",
			filename,
			err,
		)
	}

	return nil
}

func (i minioImplementation) getImageObjectName(id uint64) string {
	ids := strconv.FormatUint(id, 10)
	return strings.Join([]string{i.imagePathPrefix, ids}, "/")
}

func NewMinioImplementation(
	minioClient *minio.Client,
	bucketName string,
) (S3ImageObjectService, error) {

	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("could not check if bucket exists")
	}

	if !exists {
		if err := minioClient.MakeBucket(
			context.Background(),
			bucketName,
			minio.MakeBucketOptions{},
		); err != nil {
			return nil, fmt.Errorf(
				"failed to create bucket %s: %w",
				bucketName,
				err,
			)
		}
	}

	return minioImplementation{
		minioClient:     minioClient,
		bucketName:      bucketName,
		imagePathPrefix: "image",
	}, nil
}
