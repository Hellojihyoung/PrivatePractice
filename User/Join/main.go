package main

// 회원가입
// 1. 인증받기
// http://localhost:3000/user/join/sens/01045562725
// 2. 회원가입 하기
// http://localhost:3000/user/join
// {
//     "id" : "10",
//     "login_id" : "emily",
//     "password" : "helloworld",
//     "nickname": "우와",
//     "phone_number" : "01045562725",
//     "certification_number" : "724156"
// }

// 1. 구글 정보 가져오기
// http://localhost:3000/auth/google/login
// 2. 인증하기
// http://localhost:3000/user/join/sens/01045562725
// 3. 가입하기
// http://localhost:3000/user/join/google
// {
//     "id" : "17",
//     "password" : "helloworld",
//     "nickname": "우와",
//     "phone_number" : "01045562725",
//     "certification_number" : "829552"
// }

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

var googleOauthConfig = oauth2.Config{ 
    RedirectURL:  "http://localhost:3000/auth/google/callback", 
    ClientID:     "552627757934-dlhqnijgeajtb8spncv813eock8ug411.apps.googleusercontent.com",
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
	login_id string
	profile_img string
)

func getIdCheck(c echo.Context) error {
	params := make(map[string]string)
	c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{LoginId: ""}
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()

	var login_id string

	err = db.QueryRow("SELECT login_id FROM user WHERE login_id = ?", params["login_id"]).Scan(&login_id)

	if err != nil {
		fmt.Println(err)
	}

	response := user{LoginId: login_id}
	fmt.Println(response)
	fmt.Println(response.LoginId)

	if len(response.LoginId) != 0 {
		return errors.New("이미 존재하는 이메일 입니다")
	}

	return c.JSON(http.StatusOK, response.LoginId)
}

func compareCertificationNumber(pNum, cNum string) bool {

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var random_number string

	err = db.QueryRow("SELECT random_number FROM auth WHERE phone_number = ?", pNum).Scan(&random_number)

	if err != nil {
		fmt.Println(err)
	}

	if cNum == random_number {
		return false
	}

	return true
}

func createUser(c echo.Context) error {
	params := make(map[string]string)
	c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	// todo : 영문 대소문자, 숫자, 특수문자 중 2가지 이상 조합
	// todo : 동일문자 3회 반복 불가능
	// todo : 1234 같이 연속된 숫자, 문자 사용 불가

	// if len(params["password"]) < 10 || len(params["password"]) > 16{
	// 		return errors.New("비밀번호는 10~16자리 입니다")
	// }

	// if len(params["nickname"]) > 20{
	// 	return errors.New("닉네임은 한글 영문 숫자 특수문자 20자까지 가능합니다")
	// }

	if compareCertificationNumber(params["phone_number"], params["certification_number"]) {
		fmt.Println("인증번호 불일치")
		return errors.New("인증번호가 일치하지 않습니다")
	}

	result, err := db.Exec("INSERT INTO USER VALUES(?, ?, ?, ?, NULL, ?, 0, now())", params["id"], params["login_id"], params["password"], params["nickname"], params["phone_number"])

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}

	return c.JSON(http.StatusOK, params)
}


func createGoogleUser(c echo.Context) error {
	params := make(map[string]string)
	c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	// todo : 영문 대소문자, 숫자, 특수문자 중 2가지 이상 조합
	// todo : 동일문자 3회 반복 불가능
	// todo : 1234 같이 연속된 숫자, 문자 사용 불가

	// if len(params["password"]) < 10 || len(params["password"]) > 16{
	// 		return errors.New("비밀번호는 10~16자리 입니다")
	// }

	// if len(params["nickname"]) > 20{
	// 	return errors.New("닉네임은 한글 영문 숫자 특수문자 20자까지 가능합니다")
	// }

	if compareCertificationNumber(params["phone_number"], params["certification_number"]) {
		fmt.Println("인증번호 불일치")
		return errors.New("인증번호가 일치하지 않습니다")
	}

	result, err := db.Exec("INSERT INTO USER VALUES(?, ?, NULL, ?, ?, ?, 0, now())", params["id"], login_id, params["nickname"], profile_img, params["phone_number"])

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}

	return c.JSON(http.StatusOK, params)
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

func createAuthNum(c echo.Context) error {
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


// OAuth Join

// 로그인 요청 받을 경우 유저가 어떤 경로로 접근해서 로그인할 수 있는지 리다이렉트
func getUserByGoogle(c echo.Context) error{ 
    state := generateStateOauthCookie(c.Response())
    url := googleOauthConfig.AuthCodeURL(state) // 유저를 어떤 경로로 보내는지 URL 을 보내줌
    return c.Redirect(http.StatusTemporaryRedirect, url)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
    expiration := time.Now().Add(1 * 24 * time.Hour) // 쿠키 만료시간

    b := make([]byte, 16)
    rand.Read(b) // 랜덤
    state := base64.URLEncoding.EncodeToString(b)
    cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
    http.SetCookie(w, cookie) // 쿠키 설정
    return state
}

func googleAuthCallback(c echo.Context) error{
	r := c.Request()
	w := c.Response()
    oauthstate, _ := r.Cookie("oauthstate") // 아까 저장한 쿠키 읽어오기

    if r.FormValue("state") != oauthstate.Value { // oauthstate 값과 비교
      log.Printf("invalid google oauth state cookie:%s state:%s\n", oauthstate.Value, r.FormValue("state"))
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

   	fmt.Println(gUser)


	login_id = gUser["email"].(string)
	profile_img = gUser["picture"].(string)

	return c.JSON(http.StatusOK, string(data))
}


const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func getGoogleUserInfo(code string) ([]byte, error) {
    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
      return nil, fmt.Errorf("failed to Exchange %s", err.Error())
    }

    res, err := http.Get(oauthGoogleUrlAPI + token.AccessToken) // userinfo request by token
    if err != nil {
      return nil, fmt.Errorf("failed to Get UserInfo %s", err.Error())
    }

	// fmt.Println("---------getGoogleUserInfo--------")
	// fmt.Println(res.Body)
    return ioutil.ReadAll(res.Body)
}

func deleteUser(c echo.Context) error {
	requested_id := c.Param("id")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("DELETE from user where id = ?", requested_id)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row deleted.")
	}

	return c.JSON(http.StatusOK, result)
}

func main() {

	e := echo.New()

	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.POST("/user/join", createUser) 
	e.POST("/user/join/google", createGoogleUser) 

	e.GET("/user/join/checkId", getIdCheck)
	e.POST("/user/join/sens/:phoneNum", createAuthNum) // 휴대폰 인증
	e.DELETE("/user/delete/:id", deleteUser) 

	e.GET("/auth/google/login", getUserByGoogle) // join by google (경로를 구글 API에 이렇게 등록해놔서 그냥 둠)
   	e.GET("/auth/google/callback", googleAuthCallback)

	// e.Logger.Fatal(e.Start(":3000"))
	e.Start(":3000")

	// 구글 회원가입 getUserByGoogle -> createAuthNum -> createGoogleUser
	// 그냥 회원가입 createAuthNum -> createUser
}
