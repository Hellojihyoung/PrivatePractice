package factory02

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"

	"github.com/labstack/echo"
)

var (
	svc *s3.S3
	err = godotenv.Load()
	id = os.Getenv("AWS_ACCESS_KEY_ID")
	key = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = os.Getenv("AWS_S3_REGION")
	bucket = os.Getenv("AWS_S3_BUCKET")
	token = ""
)

func init() {
	creds := credentials.NewStaticCredentials(id, key, token)
	_, err = creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc = s3.New(session.New(), cfg)
}

func GetFactoryCertificate(c echo.Context) error{

	file := c.FormValue("file")
	err := godotenv.Load()

    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    r, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(file),
    })
 
	// url 생성
    url, err := r.Presign(2000 * time.Second)
    if err != nil {
        fmt.Println("Failed to generate a pre-signed url: ", err)
        return err
    }
 
    fmt.Println("Pre-signed URL", url)

	return c.String(http.StatusOK, url)
}