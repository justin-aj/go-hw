package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	// Add this import at the top

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var s3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
}

func main() {
	r := gin.Default()

	r.GET("/split", splitHandler)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.Run(":8080")
}

func splitHandler(c *gin.Context) {
	bucket := c.Query("bucket")
	key := c.Query("key")

	if bucket == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bucket and key query params required"})
		return
	}

	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get S3 object: %v", err)})
		return
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read body: %v", err)})
		return
	}

	text := string(body)

	// In splitHandler, after reading the file:
	numChunksStr := c.DefaultQuery("num_chunks", "3")
	numChunks, err := strconv.Atoi(numChunksStr)
	if err != nil || numChunks < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "num_chunks must be a positive integer"})
		return
	}

	words := strings.Fields(text)
	totalWords := len(words)
	chunkSize := totalWords / numChunks

	chunks := []string{}
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if i == numChunks-1 {
			end = totalWords // last chunk gets remainder
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
	}

	chunkURLs := []string{}
	for i, chunk := range chunks {
		chunkKey := fmt.Sprintf("chunks/chunk_%d.txt", i)

		_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(chunkKey),
			Body:   strings.NewReader(chunk),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to upload chunk %d: %v", i, err)})
			return
		}

		chunkURLs = append(chunkURLs, fmt.Sprintf("s3://%s/%s", bucket, chunkKey))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "split complete",
		"chunks":     chunkURLs,
		"num_chunks": len(chunks),
		"bucket":     bucket,
	})
}
