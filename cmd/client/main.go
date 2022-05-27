package main

import (
	"context"

	"github.com/aveplen-bach/s3/protos/s3file"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial(":9090", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := s3file.NewS3GatewayClient(conn)

	if _, err := client.PutImageObject(context.Background(), &s3file.ImageObject{
		Id:       1,
		Contents: []byte("hello world"),
	}); err != nil {
		panic(err)
	}
}
