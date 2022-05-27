package main

import (
	"net"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aveplen-bach/s3/internal/config"
	"github.com/aveplen-bach/s3/internal/service"
	"github.com/aveplen-bach/s3/internal/transport"
	"github.com/aveplen-bach/s3/protos/s3file"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	config, err := config.ParseFromEnv()
	if err != nil {
		zapLogger.Fatal("failed to get config", zap.Error(err))
	}

	minioClient, err := minio.New(
		config.S3Addr,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				config.S3AccessKey,
				config.S3AccessToken,
				"",
			),
			Secure: false,
		},
	)
	if err != nil {
		zapLogger.Fatal("failed to create minio client", zap.Error(err))
	}

	minioService, err := service.NewMinioImplementation(minioClient, config.S3BucketName)
	if err != nil {
		zapLogger.Fatal("failed to initialize minio-service", zap.Error(err))
	}

	grpcConn, err := net.Listen("tcp", config.GRPCListenAddr)
	if err != nil {
		zapLogger.Fatal("failed to listen to socket", zap.Error(err))
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				grpc_zap.UnaryServerInterceptor(zapLogger),
			),
		),
	)

	grpcService := transport.New(minioService)
	s3file.RegisterS3GatewayServer(server, grpcService)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(config.HTTPListenAddr, http.DefaultServeMux)
	}()

	if err := server.Serve(grpcConn); err != nil {
		zapLogger.Fatal("server failed", zap.Error(err))
	}
}
