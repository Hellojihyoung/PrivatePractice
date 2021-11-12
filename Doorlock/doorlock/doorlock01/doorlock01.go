package doorlock01

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"

	"fmt"
	"net/http"

	// "strconv"
	// "strings"
	// "time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
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

func DecryptMessage(message string) (string, error) {
	key := []byte("0123456789abcdef")
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", fmt.Errorf("could not base64 decode: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("invalid ciphertext block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

//Doorlock-2
func CheckSerialNumber(c echo.Context) error{
	params := make(map[string]string)
	c.Bind(&params)

	QRCode := params["QRCode"]
	decoded, _ := DecryptMessage(QRCode)

	slice := strings.Split(decoded, "_")
	serialNumber := slice[0]
	authCode := slice[1]

	fmt.Println(serialNumber, authCode)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var status int; // 발급해준 값
	var certCode string;

	err = db.QueryRow("SELECT door_lock_status, auth_code FROM door_lock WHERE serial_number = ? AND auth_code = ?", serialNumber, authCode).Scan(&status, &certCode)
	
	if err != nil {
		fmt.Println(err)
	}

	if status == 5 ||  status == 6 { // 인증 완료
		return errors.New("이미 인증 되었습니다")
	}

	if len(certCode) == 0 {
		return errors.New("인증 실패")
	}
	
	return c.String(http.StatusOK, certCode)
}
