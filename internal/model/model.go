package model

import "github.com/aveplen-bach/s3/protos/s3file"

type ImageObject struct {
	Id       uint64
	Contents []byte
}

func ImageObjectFromPb(in *s3file.ImageObject) ImageObject {
	return ImageObject{
		Id:       in.Id,
		Contents: in.Contents,
	}
}

func (o *ImageObject) Pb() *s3file.ImageObject {
	return &s3file.ImageObject{
		Id:       o.Id,
		Contents: o.Contents,
	}
}
