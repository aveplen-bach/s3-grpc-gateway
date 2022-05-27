package service

import (
	"context"

	"github.com/aveplen-bach/s3/internal/model"
)

type S3ImageObjectService interface {
	GetImageObject(ctx context.Context, id uint64) (model.ImageObject, error)
	PutImageObject(ctx context.Context, object model.ImageObject) error
}
