package transport

import (
	context "context"

	"github.com/aveplen-bach/s3/internal/model"
	"github.com/aveplen-bach/s3/internal/service"
	"github.com/aveplen-bach/s3/protos/s3file"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type implementation struct {
	s3file.UnimplementedS3GatewayServer

	svc service.S3ImageObjectService
}

// GetImageObject implements s3file.S3GatewayServer
func (i implementation) GetImageObject(ctx context.Context, req *s3file.GetImageObjectRequest) (*s3file.ImageObject, error) {
	object, err := i.svc.GetImageObject(ctx, req.Id)

	return object.Pb(), err
}

// PutImageObject implements s3file.S3GatewayServer
func (i implementation) PutImageObject(ctx context.Context, req *s3file.ImageObject) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, i.svc.PutImageObject(ctx, model.ImageObjectFromPb(req))
}

func New(
	svc service.S3ImageObjectService,
) s3file.S3GatewayServer {
	return implementation{
		UnimplementedS3GatewayServer: s3file.UnimplementedS3GatewayServer{},
		svc:                          svc,
	}
}
