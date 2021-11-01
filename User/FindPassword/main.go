package main

// 비밀번호 찾기
import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	user struct {
		Id   string    `json:"id"`
		LoginId   string    `json:"login_id"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		ProfileImg string `json:"profile_img"`
		PhoneNumber string `json:"phone_number"`
		LoginAttempt string `json:"login_attempt"`
		PasswordChangeDate string `json:"password_change_date"`
		Color string `json:"color"`
		CertificationNumber string `json:"certification_number"`
	}
)

type Message struct{
	To string `json:"to"`
}

type SMS struct {
	Type        string `json:"type"`
	CountryCode string `json:"countryCode"`
	From        string `json:"from"`
	Content     string `json:"content"`
	Messages    []Message `json:"messages"`
}

const (
	host     = "localhost"
	database = "HDC"
	user1     = "root"
	password = "Ant123!!!"
)

var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)

func updatePassword(c echo.Context) error{
	// {
	// 	"login_id": "",
	// 	"phone_number": "",
	// 	"certification_number": "",
	// 	"new_password": ""
	// }
	params := make(map[string]string)
    c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	if(idExist(params["login_id"])){
		fmt.Println("가입되지 않은 이메일")
		return errors.New("가입되지 않은 이메일 입니다")
	}
	
	if idExistbyPhoneNumber(params["phone_number"]){
		fmt.Println("가입되지 않은 전화번호")
		return errors.New("가입되지 않은 전화번호 입니다")
	}

	if compareCertificationNumber(params["phone_number"], params["certification_number"]){
		fmt.Println("인증번호 불일치")
		return errors.New("인증번호가 일치하지 않습니다")
	}

	result, err := db.Exec("UPDATE user set password = ? WHERE phone_num = ?", params["new_password"],  params["phone_number"])

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("update password")
    }

    return c.JSON(http.StatusOK, result)
}


func idExistbyPhoneNumber(pNum string) bool{

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var phone_number string;

	err = db.QueryRow("SELECT phone_number FROM user WHERE phone_number = ?", pNum).Scan(&phone_number)
	
	if err != nil {
		fmt.Println(err)
	}

	if len(phone_number) == 0{
		return true
	}

	return false
}

func idExist(id string) bool{

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var login_id string;

	err = db.QueryRow("SELECT login_id FROM user WHERE login_id = ?", id).Scan(&login_id)
	
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(login_id)

	if len(login_id) == 0{
		return true
	}

	return false
}

func compareCertificationNumber(pNum, cNum string) bool{

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var random_number string;

	err = db.QueryRow("SELECT random_number FROM auth WHERE phone_number = ?", pNum).Scan(&random_number)
	
	if err != nil {
		fmt.Println(err)
	}

	if cNum == random_number{
		return false
	}

	return true
}


func makeSignature(str string) string {
	secretKey := "XKOkvDVegIxAub5OOHN9z0M3dDf9mDwvGPK9aiKe"
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func verifyNumber() string {
	var randNum = []rune("0123456789")
 
	s := make([]rune, 6)
	rand.Seed(time.Now().UnixNano())
	for i := range s {
	   s[i] = randNum[rand.Intn(len(randNum))]
	}
	return string(s)
 }

func HandleMessage(phone_number string) string{
	randomNum := verifyNumber()
	request := SMS{
		Type: "SMS",
		CountryCode: "82",
		From: "01045562725",
		Content: "인증번호를 입력해 주세요 ["+ randomNum + "]",
		Messages: []Message{{
			To: phone_number,
		}},
	}

	json_data, _ := json.Marshal(request)
	reqBody := bytes.NewBuffer(json_data)

	fmt.Println("reqBody")
	fmt.Println(reqBody)

	client := &http.Client{}

	req, _ := http.NewRequest("POST", "https://sens.apigw.ntruss.com/sms/v2/services/ncp:sms:kr:273928692857:go_sns/messages", reqBody)
	fmt.Println(req.Body)
	URL := "/sms/v2/services/ncp:sms:kr:273928692857:go_sns/messages"
	accessKey := "kXvZHWWz3gIWGRoyrrld"
	timestamp := time.Now().UnixNano() / 1000000
	strTimestamp := strconv.Itoa(int(timestamp))
	sigString := "POST " + URL + "\n" + strTimestamp + "\n" + accessKey
	signature := makeSignature(sigString)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ncp-apigw-timestamp", strTimestamp)
	req.Header.Set("x-ncp-iam-access-key", accessKey)
	req.Header.Set("x-ncp-apigw-signature-v2", signature)

	res, err := client.Do(req)
	fmt.Println(res)

	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	defer res.Body.Close()

	return randomNum
}


func createAuthNum(c echo.Context) error{

	params := make(map[string]string)
    c.Bind(&params)

	requested_phoneNum := c.Param("phoneNum")
	code := HandleMessage(requested_phoneNum)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO AUTH VALUES(?, ?, ?)", 0, requested_phoneNum, code)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row inserted.")
    }

    return c.JSON(http.StatusOK, result)
}


func main() {

	e := echo.New()
  
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	  }))
	
	e.PUT("/user/findPassword", updatePassword) // 아이디 주기
	e.POST("/user/findPassword/:phoneNum", createAuthNum) // 전화번호 입력

	e.Logger.Fatal(e.Start(":3000"))
}