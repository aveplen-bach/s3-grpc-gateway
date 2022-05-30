package main

import (
	"context"
	"io/ioutil"

	"github.com/aveplen-bach/s3/protos/s3file"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial(":9090", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := s3file.NewS3GatewayClient(conn)

	filename := "/home/anon/Downloads/obama.jpg"
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	if _, err := client.PutImageObject(context.Background(), &s3file.ImageObject{
		Id:       1,
		Contents: dat,
	}); err != nil {
		panic(err)
	}
}
