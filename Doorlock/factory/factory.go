package factory

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	// "sort"
	"strconv"
	// "sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awsutil"
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

// [Factory-1]
func CreateSN(modelNumber string, year string, month string, lot string) string{

	// fmt.Println(modelNumber, year, month, lot)
	num, _ := strconv.Atoi(year)
	code := string(num - 1956)
	serialNumber := modelNumber + code + month + lot

	return serialNumber
}

func CreateSerialKey() string { // 시리얼키 생성
	var randNum = []rune("0123456789")

	s := make([]rune, 5)
	rand.Seed(time.Now().UnixNano())
	for i := range s {
		s[i] = randNum[rand.Intn(len(randNum))]
	}
	return string(s)
}

func GetUniqueKey() string { // 고유 인증키
	var randNum = []rune("0123456789")

	s := make([]rune, 5)
	rand.Seed(time.Now().UnixNano())
	for i := range s {
		s[i] = randNum[rand.Intn(len(randNum))]
	}
	return string(s)
}

func AES128(QRcode string) string{
	plaintext := []byte(QRcode)
	key := make([]byte, 32)
	rand.Read(key) // 랜덤값 키 생성
	ciphertext := make([]byte, len(plaintext))

	cip, _ := aes.NewCipher(key)
	cip.Encrypt(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

func CreateCSV(data [][]string) string{
	fileName := "result.csv"
	file, err := os.Create(fileName)
    checkError("Cannot create file", err)
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    for _, value := range data {
        err := writer.Write(value)
        checkError("Cannot write to file", err)
    }
	return fileName
}


func UploadCSV(fileName string) string{

	_, writer := io.Pipe()
	file, _ := os.Open(fileName)

	go func()  {
		gw := writer
		io.Copy(gw, file)
		file.Close()
		gw.Close()
		writer.Close()
	}()
	
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(fileName),
	}

	resp, err := svc.PutObject(params)
	if err != nil {
		fmt.Print("bad response: %s", err)
	}
	fmt.Printf("response %s", resp)
	
	return fileName
}

func CreateSerialNumber(c echo.Context) error{

	params := make(map[string]string)
    c.Bind(&params)

	modelNumber := params["modelNumber"]
	year := params["year"]
	month := params["month"]
	count, _ := strconv.Atoi(c.QueryParam("count"))

	// (2) 유효성 검증 필요
	var data = [][]string{{"no", "S/N", "QR코드"}}

	for no := 1; no <= count; no++ {
		SerialNumber := CreateSN(modelNumber, year, month, "01" + CreateSerialKey())
		QRcode := SerialNumber + "_" + GetUniqueKey()
		EncryptionQR := AES128(QRcode)
		row := []string{strconv.Itoa(no), SerialNumber, EncryptionQR}
		data = append(data, row)
		// (5) 디비 저장
    }
	// fmt.Println(data)
	URL := UploadCSV(CreateCSV(data))

	return c.String(http.StatusOK, URL)
}