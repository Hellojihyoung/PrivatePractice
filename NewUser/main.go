package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"

	"github.com/dgrijalva/jwt-go"
)

type (
	user struct {
		Id                  string `json:"id"`
		LoginId             string `json:"login_id"`
		Password            string `json:"password"`
		Nickname            string `json:"nickname"`
		ProfileImg          string `json:"profile_img"`
		PhoneNumber         string `json:"phone_number"`
		LoginAttempt        int    `json:"login_attempt"`
		PasswordChangeDate  string `json:"password_change_date"`
		Color               string `json:"color"`
		CertificationNumber string `json:"certification_number"`
	}
)

type jwtCustomClaims struct {
	Method  int      `json:"method"`
	Status  int      `json:"status"`
	UserId  string   `json:"userId"`

	jwt.StandardClaims
}
type jwtAccessClaims struct {
	Method  int      `json:"method"`
	UserId  string   `json:"userId"`
	IsTermAgree bool `json:"isTermAgree"`
	IsInfoRegistered bool `json:"isInfoRegistered"`
	IsPhoneVerified bool `json:"isPhoneVerified"`

	jwt.StandardClaims

}

type JwtClaim struct {
	Phone  string `json:"phoneNumber"`
	jwt.StandardClaims
}

type CJwtClaim struct {
	LoginedId  string `json:"loginedId"`
	LoginedPassword  string `json:"loginedPassword"`
	jwt.StandardClaims
}

const (
	host     = "localhost"
	database = "HDC"
	user1    = "root"
	password = "Ant123!!!"
)

var (
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)
	login_id          string
	profile_img       string
	rId 			  string
	rPassword 		  string
	rNickname 		  string
	rFileCount 		  string
	rImage 			  string
	rPhone 			  string
	rVerificationCode string
)

// [USER-1] 아이디 중복 확인
func getIdCheck(c echo.Context) error {
	loginId := c.QueryParam("loginId")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{LoginId: ""}
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()

	var login_id string

	err = db.QueryRow("SELECT login_id FROM user WHERE login_id = ?", loginId).Scan(&login_id)

	if err != nil {
		fmt.Println(err)
	}

	response := user{LoginId: login_id}

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

func deleteUser(c echo.Context) error {
	requested_id := c.Param("userId")

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

// USER-2
func createIdPw(c echo.Context) error {
	params := make(map[string]string)
    c.Bind(&params)

	rId = params["loginId"]
	rPassword = params["password"]

	if len(rPassword) < 10 || len(rPassword) > 16{
			return errors.New("비밀번호는 10~16자리 입니다")
	}

	registerToken, _ := generateToken(0)

	return c.JSON(http.StatusOK, registerToken)
}

func generateToken(status int) (map[string]string, error) {
	registerClaim := &jwtCustomClaims{
		0,
		status,
		rId,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 100).Unix(),
		},
	}

	registerToken := jwt.NewWithClaims(jwt.SigningMethodHS256, registerClaim)
	rt, err := registerToken.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		// return err
	}

	return map[string]string{
		"register_token" : rt,
	}, nil
}

func verifyToken(r *http.Request) bool{
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Claims, err)

	user := jwtCustomClaims{}
	token, err = jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Valid, user.UserId, err)

	if user.UserId == rId{
		fmt.Print("-----------접근 가능 ------------")
		return false
	}

	return true
}

// USER-3
func registerTermsAcception(c echo.Context) error {

	return nil
}

// USER-4

func createUserInfo(c echo.Context) error {
	if verifyToken(c.Request()) {
		return errors.New("접근 불가능!!!!!!!!!!!!!")
	}
	nickname := c.FormValue("nickname")
	fileCount := c.FormValue("fileCount")
	image := c.FormValue("image") // 이것도 FormFile 형태로 받는다.

	rNickname = nickname
	rFileCount = fileCount
	rImage = image // 사실 여기는 s3버킷 이미지 경로를 넣어준다.

	registerToken, _ := generateToken(2)

	return c.JSON(http.StatusOK, registerToken)
}

// USER-5
func createUserPhone(c echo.Context) error {
	params := make(map[string]string)
    c.Bind(&params)

	phone := params["phone"]
	verificationCode := params["verificationCode"]

	rPhone = phone
	rVerificationCode = verificationCode

	if compareCertificationNumber(phone, verificationCode) {
		fmt.Println("인증번호 불일치")
		return errors.New("인증번호가 일치하지 않습니다")
	}

	registerToken, _ := generateToken(3)

	return c.JSON(http.StatusOK, registerToken)
}

// USER-6
func generateSuccessToken(id string, password string) (map[string]string, error) {
	accessClaims := &CJwtClaim{
		id,
		password,
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
		id,
		password,
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

	return map[string]string{
		"access_token" : at,
		"refresh_token" : rt,
	}, nil
}

func getUser(c echo.Context) error{
	params := make(map[string]string)
    c.Bind(&params)
	
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{LoginId: "", Password: "", LoginAttempt: 0}
			return c.JSON(http.StatusInternalServerError, response)
	}
	
	defer db.Close()
	
	var login_id string;
	var password string;
	var login_attempt int;

	err = db.QueryRow("SELECT login_id, password, login_attempt FROM user WHERE login_id = ?", params["loginId"]).Scan(&login_id, &password, &login_attempt)
	
	if err != nil {
		fmt.Println(err)
	}
	response := user{LoginId: login_id, Password: password, LoginAttempt: login_attempt}

	if len(response.LoginId) == 0{
		login_attempt ++
		return echo.NewHTTPError(login_attempt)
	}

	if (params["password"] != response.Password){
		login_attempt ++
		return echo.NewHTTPError(login_attempt)
	}

	token, _:= generateSuccessToken(response.LoginId, response.Password)

	
	// fmt.Println((response.PasswordChangeDate))

	// compareWith := time.Now().AddDate(0, -3, 0)
	// before := compareWith.Format("2006-01-02 15:04:05")
	// fmt.Println(before) 
	
	// arr := []string{before, response.PasswordChangeDate}
	// sl := sort.StringSlice(arr)
	// sl.Sort()
	// fmt.Println(sl)

	// if sl[0] == response.PasswordChangeDate{
	// 	fmt.Println("update password") // updatePassword Login is in FindPassword/main.go
	// }

	return c.JSON(http.StatusOK, token)
}

// User-7
func findId(c echo.Context) error{
	r := c.Request()
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Claims, err)

	find := JwtClaim{}
	token, err = jwt.ParseWithClaims(tokenString, &find, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	phoneNum := find.Phone

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{LoginId: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()

	err = db.QueryRow("SELECT login_id FROM user WHERE phone_number = ?", phoneNum).Scan(&login_id)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := user{LoginId: login_id}

	if len(response.LoginId) == 0{
		return errors.New("가입되지 않은 이메일 입니다")
	}

	return c.JSON(http.StatusOK, response.LoginId)
}

// User-8
func findPassword(c echo.Context) error{
	r := c.Request()
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Claims, err)

	find := JwtClaim{}
	token, err = jwt.ParseWithClaims(tokenString, &find, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	phoneNum := find.Phone

	params := make(map[string]string)
    c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
			return err
	}

	defer db.Close()

	result, err := db.Exec("UPDATE user set password = ? WHERE phone_num = ?", params["password"],  phoneNum)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("update password")
    }

	return c.JSON(http.StatusOK, "update Password")
}


// User-9
func decodeToken(r *http.Request) string{
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Claims, err)

	user := jwtAccessClaims{}
	token, err = jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Valid, user.UserId, err)

	return user.UserId
}

func getUserInfo(c echo.Context) error{ 
	userId := decodeToken(c.Request())
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{Nickname: "", ProfileImg: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}

	defer db.Close()
	
	var nickname string;
	var profile_img string;

	err = db.QueryRow("SELECT nickname, profile_img FROM user WHERE login_id = ?", userId).Scan(&nickname, &profile_img)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := user{Nickname: nickname, ProfileImg: profile_img}

	return c.JSON(http.StatusOK, response)
}

func updateUser(c echo.Context) error{
	userId := c.Param("userId")
	profileImg := c.FormValue("profileImg")
	nickname := c.FormValue("nickname")
	// profileImage := c.FormValue("profileImage")
	fileCount := c.FormValue("fileCount")

	params := make(map[string]string)
    c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()


	result, err := db.Exec("UPDATE FAQ set nickname = ?, profile_img = ?, fileCount =?, WHERE id = ?", nickname, profileImg, fileCount, userId)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row content updated.")
    }

    return c.JSON(http.StatusOK, result)
}

func main() {

	e := echo.New()

	e.GET("/api/v1/user/register/login-id", getIdCheck) // USER-1
	e.POST("/api/v1/user/register/login-id", createIdPw) // USER-2
	e.POST("/api/v1/user/register/terms", registerTermsAcception) // USER-3
	e.POST("/api/v1/user/register/info", createUserInfo) // USER-4
	e.POST("/api/v1/user/register/phone", createUserPhone) // USER-5
	e.POST("/api/v1/user/login", getUser) // USER-6

	e.POST("api/v1/user/login-id/find", findId) // USER-7
	e.POST("api/v1/user/password", findPassword) // USER-8

	e.GET("/api/v1/user/:userId", getUserInfo) // USER-9
	e.PUT("api/v1/user/:userId", updateUser)
	e.DELETE("/api/v1/user/:userId", deleteUser) 

	e.Start(":3000")
}
