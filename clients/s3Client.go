package client

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func S3client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("BUCKETS_LOCATION")))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = os.Getenv("BUCKETS_LOCATION")
		o.UseAccelerate = false
	})
	return client, nil
}
