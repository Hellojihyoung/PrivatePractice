package factory

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"database/sql"
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
	"github.com/aws/aws-sdk-go/aws/awsutil"
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

const (
	host     = "localhost"
	database = "HDC"
	user1    = "root"
	password = "Ant123!!!"
)

var (	
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)
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

func CreateSN(modelNumber string, year string, month string, lot string) string{
	num, _ := strconv.Atoi(year)
	code := string(num - 1956)
	serialNumber := modelNumber + code + month + lot

	return serialNumber
}

func CreateSerialKey() string { // 시리얼키 생성
	var randNum = []rune("0123456789")

	s := make([]rune, 5)
	rand.Seed(time.Now().UnixNano())
	time.Sleep(10)
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
	byteMsg := []byte(QRcode)
	key := []byte("0123456789abcdef")
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Errorf("could not create new cipher: %v", err)
		return ""
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(crand.Reader, iv); err != nil {
		fmt.Errorf("could not encrypt: %v", err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)
	
	return base64.StdEncoding.EncodeToString(cipherText)
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

func CreateCSV(data [][]string) string{
	fileName := "output.csv"
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

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Print(err)
	}
	defer file.Close()

	// fileInfo, _ := file.Stat()
	// size := fileInfo.Size()
	// buffer := make([]byte, size)

	// _, s3err := svc.PutObject(&s3.PutObjectInput{
	// 	Bucket:               aws.String(bucket),
	// 	Key:                  aws.String(fileName),
	// 	ACL:                  aws.String("private"),
	// 	Body:                 bytes.NewReader(buffer),
	// 	ContentLength:        aws.Int64(size),
	// 	// ContentType:          aws.String(http.DetectContentType(buffer)),
	// 	ContentDisposition:   aws.String("attachment"),
	// 	ServerSideEncryption: aws.String("AES256"),
	// 	StorageClass:         aws.String("INTELLIGENT_TIERING"),
	// })
	// fmt.Println(s3err)

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	// path := file.Name()
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(file.Name()),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	resp, err := svc.PutObject(params)

	fmt.Printf("response %s", awsutil.StringValue(resp))

	return fileName
}

func InsertDB(modelNumber string, serialNumber string, authCode string){
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO door_lock VALUES(?, NULL, ?, ?, NULL, NULL, NULL, ?, NULL, NULL, NULL, NULL, 0, NULL, NULL, NULL, NULL, NULL, NULL, NOW(), NULL)", 0, modelNumber, serialNumber, authCode)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}
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
		AuthCode := GetUniqueKey()
		QRcode := SerialNumber + "_" + AuthCode
		EncryptionQR := AES128(QRcode)
		row := []string{strconv.Itoa(no), SerialNumber, EncryptionQR}
		data = append(data, row)
		InsertDB(modelNumber, SerialNumber, AuthCode)
    }
	// fmt.Println(data)
	URL := UploadCSV(CreateCSV(data))

	return c.String(http.StatusOK, URL)
}