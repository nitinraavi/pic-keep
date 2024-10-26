package controllers

import (
	"bytes"
	"context"
	"fmt"
	client "go-jwt/clients"
	"go-jwt/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// type HttpRequester struct{}

//	func (httpReq HttpRequester) Get(url string) (resp *http.Response, err error) {
//		return http.Get(url)
//	}
//
//	func (httpReq HttpRequester) Put(url string, contentLength int64, body io.Reader) (resp *http.Response, err error) {
//		putRequest, err := http.NewRequest("PUT", url, body)
//		if err != nil {
//			return nil, err
//		}
//		putRequest.ContentLength = contentLength
//		return http.DefaultClient.Do(putRequest)
//	}
//
//	func (httpReq HttpRequester) Delete(url string) (resp *http.Response, err error) {
//		delRequest, err := http.NewRequest("DELETE", url, nil)
//		if err != nil {
//			return nil, err
//		}
//		return http.DefaultClient.Do(delRequest)
//	}
func UploadImages(c *gin.Context) {
	// Get the user details from request
	// httpRequester := &HttpRequester{}
	// Intiate S3Client
	s3client, err := client.S3client()
	if err != nil {
		log.Println("error while configuring the client", err)
	}
	var bucketSize int64
	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User details not found"})
		return
	}

	userData := user.(models.User)

	// Parse the multipart form
	if err := c.Request.ParseMultipartForm(32); err != nil { // 10 MB limit
		c.String(http.StatusBadRequest, "Error parsing form: %s", err.Error())
		return
	}
	// Get the file from the request to Upload
	files := c.Request.MultipartForm.File["file"]

	// Check the number of files
	if len(files) > 10 {
		c.JSON(http.StatusBadRequest, "Exceeded maximum number of files (10)")
		return
	}

	paginator := s3.NewListObjectsV2Paginator(s3client, &s3.ListObjectsV2Input{
		Bucket: &userData.BucketName,
	})

	for paginator.HasMorePages() {

		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			fmt.Println("Error listing objects:", err)
			os.Exit(1)
		}

		for _, obj := range page.Contents {

			size := *obj.Size // Dereference the pointer to get the int64 value
			bucketSize += size

		}
	}
	fmt.Println(bucketSize)
	if bucketSize > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Your Subscription exceeds the size limit of 1GB. Please choose the Next Subscription."})
		return
	}
	// var filenames string
	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Couldn't open thefiles: %w", err)
		}
		defer file.Close()
		// Read the filename from the file header
		filenames := fileHeader.Filename

		// Read the file content
		fileContent, err := io.ReadAll(file)
		if err != nil {
			log.Println("Couldn't read the file content:", err)
			continue
		}

		// Check the file size
		if int64(len(fileContent)) > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("File %s exceeds the size limit of 10MB. Please choose a smaller file.", filenames)})
			continue
		}

		// Checking the content type
		// so we don't allow files other than images
		filetype := http.DetectContentType(fileContent)
		if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/jpg" {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Provided file %s format is not allowed. Please upload a JPEG, JPG, or PNG image.", filenames)})
			continue
		}

		// // Check file size
		// fileInfo, err := os.Stat(filenames)
		// if err != nil {
		// 	panic(err)
		// }
		// if fileInfo.Size() > 10*1024*1024 {
		// 	log.Printf("File %s exceeds the size limit of 10MB. Please choose a smaller file.", fileHeader.Filename)
		// 	continue
		// }
		// buff := make([]byte, 512)
		// _, err = file.Read(buff)
		// if err != nil {
		// 	log.Println("Couldn't Read the buffer: %w", err)
		// 	break
		// }

		// // checking the content type
		// // so we don't allow files other than images
		// filetype := http.DetectContentType(buff)
		// if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/jpg" {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"message": fmt.Sprintf("provided file %s format is not allowed. Please upload a JPEG,JPG or PNG image ", fileHeader.Filename),
		// 	})
		// 	break
		// }
		// _, err = file.Seek(0, io.SeekStart)
		// if err != nil {
		// 	log.Println("Cannt find the first file", err)
		// 	break
		// }
		// err = os.MkdirAll("./uploads", os.ModePerm)
		// if err != nil {
		// 	log.Println("Cannt create the Folder", err)
		// 	break
		// }

		// f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		// if err != nil {
		// 	log.Println("Cannt create the Files in the folder", err)
		// 	break
		// }

		// defer f.Close()

		// _, err = io.Copy(f, file)
		// if err != nil {
		// 	log.Println("Cannt copy the Files in the folder", err)
		// 	break
		// }

		// Create a new buffer to store the request body
		// body := bytes.NewReader(fileContent)
		// // Create a new multipart writer
		// writer := multipart.NewWriter(body)
		// defer writer.Close()

		// _, err = file.Seek(0, io.SeekStart)
		// if err != nil {
		// 	log.Println("Cannt find the first file", err)
		// 	break
		// }

		// part, err := writer.CreateFormFile("files", fileHeader.Filename)
		// if err != nil {
		// 	log.Panic("Can't create the File: %w", err)
		// }

		// _, err = io.Copy(part, file)
		// if err != nil {
		// 	log.Panic("Can't copy the File: %w", err)
		// }

		// Upload the file to S3
		_, err = s3client.PutObject(c, &s3.PutObjectInput{
			Bucket:      aws.String(userData.BucketName),
			Key:         aws.String(fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(filenames))),
			Body:        bytes.NewReader(fileContent),
			ContentType: aws.String(filetype),
		})
		if err != nil {
			log.Println(err)
		}
		// presignClient := s3.NewPresignClient(s3client)
		// presigner := Presigner{PresignClient: presignClient}
		// bucketKey := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
		// presignedPutRequest, err := presigner.PutObject(userData.BucketName, bucketKey, 60)
		// if err != nil {
		// 	log.Println("Couldn't Generate PreSigned URL: %w", err)
		// }
		// // Create a new HTTP request with the specified URL
		// req, err := http.NewRequest("PUT", presignedPutRequest.URL, body)
		// if err != nil {
		// 	panic(err)
		// }

		// // Set the content type
		// req.Header.Set("Content-Type", writer.FormDataContentType())
		// // Use the HttpRequester to send the request
		// resp, err := httpRequester.Put(presignedPutRequest.URL, int64(body.Len()), body)
		// if err != nil {
		// 	panic(err)
		// }
		// log.Printf("Response: %v", resp)

	}

}
