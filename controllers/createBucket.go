package controllers

import (
	"context"
	"fmt"
	client "go-jwt/clients"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func CreateBucket(userIdentifier string) (string, error) {
	// form the user name as bucket name
	bucketName := fmt.Sprintf("user-%s-bucket", userIdentifier)
	// Create bucket with the user name
	s3client, err := client.S3client()
	if err != nil {
		fmt.Println("Error creating S3 client:", err)
		return "", err

	}
	// Create the bucket
	_, err = s3client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: &bucketName,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(os.Getenv("BUCKETS_LOCATION")),
		},
	})
	if err != nil {
		log.Println("Couldn't create bucket: %w", err)

	}
	return bucketName, nil
}
