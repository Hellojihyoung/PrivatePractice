package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type (
	user struct {
		Id                  string `json:"id"`
		LoginId             string `json:"login_id"`
		Password            string `json:"password"`
		Nickname            string `json:"nickname"`
		ProfileImg          string `json:"profile_img"`
		PhoneNumber         string `json:"phone_number"`
		LoginAttempt        string `json:"login_attempt"`
		PasswordChangeDate  string `json:"password_change_date"`
		Color               string `json:"color"`
		CertificationNumber string `json:"certification_number"`
	}
)

type Message struct {
	To string `json:"to"`
}

type SMS struct {
	Type        string    `json:"type"`
	CountryCode string    `json:"countryCode"`
	From        string    `json:"from"`
	Content     string    `json:"content"`
	Messages    []Message `json:"messages"`
}

type (
	stateInfo struct {
		Id   		  string    `json:"id"`
		ServiceName   string    `json:"service_name"`
		State 		  string 	`json:"state"`
	}
)

type JwtClaim struct {
	Phone  string `json:"phoneNumber"`
	jwt.StandardClaims
}

type CJwtClaim struct {
	code  string `json:"code"`
	jwt.StandardClaims
}

type VerificationToken struct {
	verificationToken  string `json:"verificationToken"`
	verificationTokenExpires int64 `json:"verificationTokenExpires"`
}

var googleOauthConfig = oauth2.Config{ 
    RedirectURL:  "http://localhost:3000/auth/google/callback", 
    ClientID:     "884159187834-3mv80c26ad3jiuvn1runqrht5f4ji66t.apps.googleusercontent.com",
    ClientSecret: "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",
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
    Endpoint:     google.Endpoint,
}

const (
	host     = "localhost"
	database = "HDC"
	user1    = "root"
	password = "Ant123!!!"
)

var (
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)
	serverState string
	codeForm string
	refresh_token_ string
)

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

func HandleMessage(phone_number string) string {
	randomNum := verifyNumber()
	request := SMS{
		Type:        "SMS",
		CountryCode: "82",
		From:        "01045562725",
		Content:     "인증번호를 입력해 주세요 [" + randomNum + "]",
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

func idExist(pNum string, Id string) bool{

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var phone_number string;

	err = db.QueryRow("SELECT phone_number FROM user WHERE phone_number = ? AND login_id = ?", pNum, Id).Scan(&phone_number)
	
	if err != nil {
		fmt.Println(err)
	}

	if len(phone_number) != 0{
		return false
	}

	return true
}

func createNum(c echo.Context) error {
	params := make(map[string]string)
    c.Bind(&params)

	requested_phoneNum := params["phoneNumber"]
	requested_loginId := params["loginId"]
	requested_type := c.QueryParam("type")

	if requested_type == "password"{ // 비밀번호 찾기일 경우
		if idExist(requested_phoneNum, requested_loginId){
			fmt.Println("가입되지 않은 전화번호")
			return errors.New("가입되지 않은 전화번호 입니다")
		}
	}

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

	// expirationTime := time.Now().Add(time.Second * 180)
	expirationTime := 180

	return c.JSON(http.StatusOK, expirationTime)
}

func compareCertificationNumber(pNum, code string) bool{

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var random_number string; // 발급해준 값

	err = db.QueryRow("SELECT random_number FROM auth WHERE phone_number = ?", pNum).Scan(&random_number)
	
	if err != nil {
		fmt.Println(err)
	}

	if code == random_number{ // 입력값 == 발급값 (db 안의 값)
		return false
	}

	return true
}

func generateVerificationToken(phoneNumber string) (map[string]string, error){
	claims := &JwtClaim{
		phoneNumber,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		// return err
	}

	return map[string]string{
		"verificationToken": t,
		"verificationTokenExpires": string(claims.ExpiresAt),
	}, nil

}

func confirmationNumber(c echo.Context) error {

	params := make(map[string]string)
    c.Bind(&params)

	requested_phoneNumber := params["phoneNumber"]
	requested_code := params["code"]

	if compareCertificationNumber(requested_phoneNumber, requested_code){
		fmt.Println("인증번호 불일치")
		return errors.New("인증번호가 일치하지 않습니다")
	}

	verificationToken, _ := generateVerificationToken(requested_phoneNumber)

	return c.JSON(http.StatusOK, verificationToken)
}

// 로그인 요청 받을 경우 유저가 어떤 경로로 접근해서 로그인할 수 있는지 리다이렉트
func getUserByGoogle(c echo.Context) error{ 
	serviceName := c.Param("serviceName")
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()
	
	var state string;

	err = db.QueryRow("SELECT state FROM state WHERE service_name = ?", serviceName).Scan(&state)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := stateInfo{State: state}

    url := googleOauthConfig.AuthCodeURL(response.State) // 유저를 어떤 경로로 보내는지 URL 을 보내줌
	serverState = state
    return c.Redirect(http.StatusTemporaryRedirect, url)
}


func googleAuthCallback(c echo.Context) error{
	r := c.Request()
	w := c.Response()
	codeForm = r.FormValue("code")
    if r.FormValue("state") != serverState { // oauthstate 값과 비교
      log.Printf("invalid google oauth state cookie:%s state:%s\n", serverState, r.FormValue("state"))
      http.Redirect(w, r, "/", http.StatusTemporaryRedirect) // 다르면 redirect
      return errors.New("invalid google oauth state cookie")
    }

    data, err := getGoogleUserInfo(r.FormValue("code")) // userinfo 가져오기
    if err != nil {
      log.Println(err.Error())
      http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
      return errors.New("wrong user")
    }
	
	// userinfo
	gUser := make(map[string]interface{})

   	if err := json.Unmarshal(data, &gUser); err != nil {
      fmt.Println(err.Error())
   }
   	fmt.Println("code")
   	fmt.Println(r.FormValue("code"))
   	customToken, _ := generateToken(r.FormValue("code"))
	fmt.Println(customToken["access_token"])

	return c.String(http.StatusOK, string(data))
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v3/userinfo?access_token="

func generateToken(code string) (map[string]string, error) {
	accessClaims := &CJwtClaim{
		code,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		// return err
	}

	refreshClaims := &CJwtClaim{
		code,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2160).Unix(), // 3개월
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, err := refreshToken.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		// return err
	}

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO refresh_token VALUES(?, ?, ?)", 0, 1, rt)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}
	
	refresh_token_ = rt

	return map[string]string{
		"access_token" : at,
		"refresh_token" : rt,
	}, nil
}

func getGoogleUserInfo(code string) ([]byte, error) {

    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
      return nil, fmt.Errorf("failed to Exchange %s", err.Error())
    }
	
	refresh_token_ = token.RefreshToken

	// fmt.Println("Access Token")
	// fmt.Println(token.AccessToken)
	// fmt.Println("Refresh Token")
	
	// fmt.Println(token.RefreshToken)
	// fmt.Println(refresh_token_)

    res, err := http.Get(oauthGoogleUrlAPI + token.AccessToken) // userinfo request by token
    if err != nil {
      return nil, fmt.Errorf("failed to Get UserInfo %s", err.Error())
    }

    return ioutil.ReadAll(res.Body)
}

func createState(c echo.Context) error{
	
	serviceName := c.QueryParam("serviceName")
	fmt.Println(serviceName)
	b := make([]byte, 16)
    rand.Read(b) 
    state := base64.URLEncoding.EncodeToString(b)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO STATE VALUES(?, ?, ?)", 0, serviceName, state)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}

    return c.String(http.StatusOK, state)
}

func renewalToken(c echo.Context) ( error) {
	
	accessClaims := &CJwtClaim{
		codeForm,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  at,
	  })
}

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.POST("/api/v1/auth/phone-verification", createNum) // AUTH-1
	e.POST("/api/v1/auth/phone-verification/verify", confirmationNumber) // AUTH-2

	e.POST("/api/v1/auth/state", createState) // AUTH-3
	e.GET("/api/v1/auth/redirect/:serviceName", getUserByGoogle) // AUTH-4
   	e.GET("/auth/google/callback", googleAuthCallback)

	e.POST("/api/v1/auth/refresh-token", renewalToken) // AUTH-5

	e.Start(":3000")
}