package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCListenAddr string
	HTTPListenAddr string

	S3Addr        string
	S3AccessToken string
	S3AccessKey   string
	S3BucketName  string
}

func missingKey(key string) error {
	return fmt.Errorf(
		"missing key %s in env",
		key,
	)
}

const GRPCListenAddrKey = "GRPC_LISTEN_ADRR"
const HTTPListenAddrKey = "HTTP_LISTEN_ADRR"
const S3AddrKey = "S3_ADDR"
const S3AccessTokenKey = "S3_ACCESS_TOKEN"
const S3AccessKeyKey = "S3_ACCESS_KEY"
const S3BucketNameKey = "S3_BUCKET_NAME"

func ParseFromEnv() (Config, error) {
	viper.AutomaticEnv()

	config := Config{}

	keys := []string{
		GRPCListenAddrKey,
		HTTPListenAddrKey,
		S3AddrKey,
		S3AccessTokenKey,
		S3AccessKeyKey,
		S3BucketNameKey,
	}

	for _, v := range keys {
		if !viper.IsSet(v) {
			return config, missingKey(v)
		}
	}

	return Config{
		GRPCListenAddr: viper.GetString(GRPCListenAddrKey),
		HTTPListenAddr: viper.GetString(HTTPListenAddrKey),
		S3Addr:         viper.GetString(S3AddrKey),
		S3AccessToken:  viper.GetString(S3AccessTokenKey),
		S3AccessKey:    viper.GetString(S3AccessKeyKey),
		S3BucketName:   viper.GetString(S3BucketNameKey),
	}, nil
}
