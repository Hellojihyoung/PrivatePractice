package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo"
)

const (
	PartSize   = 50_000_000
	RETRIES    = 2
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

type partUploadResult struct {
	completedPart *s3.CompletedPart
	err           error
}

func init() {
	creds := credentials.NewStaticCredentials(id, key, token)
	_, err = creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc = s3.New(session.New(), cfg)
}


var wg = sync.WaitGroup{}
// var ch = make(chan partUploadResult)
var ch chan partUploadResult

// // by multipart
func uploadToS3(resp *s3.CreateMultipartUploadOutput, fileBytes []byte, partNum int, wg *sync.WaitGroup) {
	ch = make(chan partUploadResult) 
	defer wg.Done()
	var try int

	fmt.Printf("Uploading %v \n", len(fileBytes))
	for try <= RETRIES {
		uploadRes, err := svc.UploadPart(&s3.UploadPartInput{
			Body:          bytes.NewReader(fileBytes),
			Bucket:        resp.Bucket,
			Key:           resp.Key,
			PartNumber:    aws.Int64(int64(partNum)),
			UploadId:      resp.UploadId,
			ContentLength: aws.Int64(int64(len(fileBytes))),
		})
		if err != nil {
			fmt.Println(err)
			if try == RETRIES {
				ch <- partUploadResult{nil, err}
				return
			} else {
				try++
				time.Sleep(time.Duration(time.Second * 15))
			}
		} else {
			ch <- partUploadResult{
				&s3.CompletedPart{
					ETag:       uploadRes.ETag,
					PartNumber: aws.Int64(int64(partNum)),
				}, nil,
			}
			return
		}
	}
	ch <- partUploadResult{}
}

func uploadImage(c echo.Context) error{
	
	file, err := c.FormFile("img")
	if err != nil {
		fmt.Print(err)
		return err
	}

	src, err := file.Open()
	if err!=nil{
		return err
	}
	defer src.Close()

	fmt.Println(file.Filename)

	dst, _ := os.Create(file.Filename)
	defer os.Remove(file.Filename)

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	stat, _ := dst.Stat()
	fileSize := stat.Size()


	buffer := make([]byte, fileSize)

	_, _ = dst.Read(buffer)

	expiryDate := time.Now().AddDate(0, 0, 1)

	createdResp, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(file.Filename),
		Expires: &expiryDate,
	})

	if err != nil {
		fmt.Print(err)
		return err
	}

	var start, currentSize int
	var remaining = int(fileSize)
	var partNum = 1
	var completedParts []*s3.CompletedPart
	for start = 0; remaining > 0; start += PartSize {
		wg.Add(1)
		if remaining < PartSize {
			currentSize = remaining
		} else {
			currentSize = PartSize
		}
		go uploadToS3(createdResp, buffer[start:start+currentSize], partNum, &wg)

		remaining -= currentSize
		fmt.Printf("Uplaodind of part %v started and remaning is %v \n", partNum, remaining)
		partNum++

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		if result.err != nil {
			_, err = svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(dst.Name()),
				UploadId: createdResp.UploadId,
			})
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		}
		fmt.Printf("Uploading of part %v has been finished \n", *result.completedPart.PartNumber)
		completedParts = append(completedParts, result.completedPart)
	}

	// Ordering the array based on the PartNumber as each parts could be uploaded in different order!
	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	// Signalling AWS S3 that the multiPartUpload is finished
	resp, err := svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   createdResp.Bucket,
		Key:      createdResp.Key,
		UploadId: createdResp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})

	if err != nil {
		fmt.Print(err)
		return err
	} 
	location, _ := json.Marshal(resp.Location)
	fmt.Println(string(location))
		
	// return c.String(http.StatusOK, resp.String()) 
	return c.String(http.StatusOK, string(location)) 
}

// Pipe 써서 한것,,
// func uploadImage(c echo.Context) error{
// 	img, err := c.FormFile("file")
// 	if err != nil {
// 		fmt.Print(err)
// 		return err
// 	}

// 	_, writer := io.Pipe()
// 	file, _ := img.Open()

// 	go func()  {
// 		gw := writer
// 		io.Copy(gw, file)
// 		file.Close()
// 		gw.Close()
// 		writer.Close()
// 	}()
	
// 	params := &s3.PutObjectInput{
// 		Bucket: aws.String(bucket),
// 		Key: aws.String(img.Filename),
// 	}

// 	resp, err := svc.PutObject(params)
// 	if err != nil {
// 		fmt.Print("bad response: %s", err)
// 	}
// 	fmt.Printf("response %s", awsutil.StringValue(resp))
	
// 	return c.JSON(http.StatusOK, "uploaded" + img.Filename)
// }

func downloadImage(c echo.Context) error {
	img := c.FormValue("img")
	err := godotenv.Load()

    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    r, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
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
        Region: aws.String(region)},
    )

    downloader := s3manager.NewDownloader(sess)

    numBytes, err := downloader.Download(file,
        &s3.GetObjectInput{
            Bucket: aws.String(bucket),
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

