package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/labstack/echo"
)

func initEnv() map[string]string{

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
 
	env := map[string]string{
		"id": os.Getenv("AWS_ACCESS_KEY_ID"),
		"key" : os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"region" : os.Getenv("AWS_S3_REGION"),
		"bucket" : os.Getenv("AWS_S3_BUCKET"),
		"token" : "",
	}

	return env
}

func init(){
	fmt.Println("dffdf")
	
}

func uploadImage(c echo.Context) error {
	img := c.FormValue("img")

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env := initEnv()

	creds := credentials.NewStaticCredentials(env["id"],env["key"], env["token"])

	_, err = creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}

	cfg := aws.NewConfig().WithRegion(env["region"]).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	file, err := os.Open(img)

	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	path := file.Name()
	params := &s3.PutObjectInput{
		Bucket:        aws.String(env["bucket"]),
		Key:           aws.String(path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	resp, err := svc.PutObject(params)
	if err != nil {
		fmt.Printf("bad response: %s", err)
	}

	fmt.Printf("response %s", awsutil.StringValue(resp))

	return c.String(http.StatusOK, awsutil.StringValue(resp))
}

func downloadImage(c echo.Context) error {
	img := c.FormValue("img")
	err := godotenv.Load()

    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

 	env := initEnv()

    if env["bucket"] == "" {
        log.Fatal("an s3 bucket was unable to be loaded from env vars")
    }

 	creds := credentials.NewStaticCredentials(env["id"], env["key"], env["token"])
	_, err = creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
		return err
	}

	cfg := aws.NewConfig().WithRegion(env["region"]).WithCredentials(creds)

	svc := s3.New(session.New(), cfg)
	fmt.Println(reflect.TypeOf(svc))
    r, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(env["bucket"]),
        Key:    aws.String(img),
    })
 
	// url 생성
    url, err := r.Presign(15 * time.Second)
    if err != nil {
        fmt.Println("Failed to generate a pre-signed url: ", err)
        return err
    }
 
    fmt.Println("Pre-signed URL", url)

	slice := strings.Split(url, "/")
	imageKey := strings.Split(slice[3], "?")[0]
	fmt.Println(imageKey)

    file, err := os.Create(imageKey)
    if err != nil {
        fmt.Println(err)
		return err
    }

    defer file.Close()

    // Initialize a session in us-west-2 that the SDK will use to load
    // credentials from the shared credentials file ~/.aws/credentials.
    sess, _ := session.NewSession(&aws.Config{
        Region: aws.String(env["region"])},
    )

    downloader := s3manager.NewDownloader(sess)

    numBytes, err := downloader.Download(file,
        &s3.GetObjectInput{
            Bucket: aws.String(env["bucket"]),
            Key:    aws.String(imageKey),
        })
    if err != nil {
        fmt.Println(err)
		return err

    }

    fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return c.String(http.StatusOK, "Downloaded" + file.Name())
}

func main() {
	e := echo.New()

	e.POST("/image", uploadImage)
	e.GET("/image", downloadImage)

	e.Logger.Fatal(e.Start(":3000"))
}
