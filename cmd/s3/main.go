package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

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
	// ============================== logger ==================================
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// ============================== config ==================================
	config, err := config.ParseFromEnv()
	if err != nil {
		zapLogger.Fatal("failed to get config", zap.Error(err))
	}

	// =========================== minio client ===============================
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

	// =========================== minio service ==============================
	minioService, err := service.NewMinioImplementation(minioClient, config.S3BucketName)
	if err != nil {
		zapLogger.Fatal("failed to initialize minio-service", zap.Error(err))
	}

	grpcConn, err := net.Listen("tcp", config.GRPCListenAddr)
	if err != nil {
		zapLogger.Fatal("failed to listen to socket", zap.Error(err))
	}

	// ============================ grpc server ===============================
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

	// ============================ health live ===============================
	router := gin.Default()
	router.GET("/s3g/health/live", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// =============================== start ==================================
	if err := server.Serve(grpcConn); err != nil {
		zapLogger.Fatal("server failed", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    ":8082",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			zapLogger.Info("listening "+srv.Addr, zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("server forced to shutdown: ", zap.Error(err))
	}

	zapLogger.Info("server exiting")
}
