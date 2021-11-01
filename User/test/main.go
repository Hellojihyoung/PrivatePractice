package main

// login
// http://localhost:3000/auth/google/login
import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	}
)

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
	user1     = "root"
	password = "Ant123!!!"
)

var (
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)
	email string
)

func getUser(c echo.Context) error{ // 로그인
	params := make(map[string]string)
    c.Bind(&params)
	
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{Id: "", LoginId: "", Password: "", Nickname: "", ProfileImg: "", PhoneNumber: "", LoginAttempt: "", PasswordChangeDate: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()
	
	var id string;
	var login_id string;
	var password string;
	var nickname string;
	var profile_img string;
	var phone_number string;
	var login_attempt string;
	var password_change_date string;

	err = db.QueryRow("SELECT * FROM user WHERE login_id = ?", params["login_id"]).Scan(&id, &login_id, &password, &nickname, &profile_img, &phone_number, &login_attempt, &password_change_date)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := user{Id: id, LoginId: login_id, Password: password, Nickname: nickname, ProfileImg: profile_img, PhoneNumber: phone_number, LoginAttempt: login_attempt, PasswordChangeDate: password_change_date}
	fmt.Println(response)
	fmt.Println(response.Id)

	if len(response.Id) == 0{
		return errors.New("가입되지 않은 아이디입니다")
	}

	if (params["password"] != response.Password){
		return errors.New("비밀번호가 일치하지 않습니다")
	}

	
	fmt.Println((response.PasswordChangeDate))

	compareWith := time.Now().AddDate(0, -3, 0)
	before := compareWith.Format("2006-01-02 15:04:05")
	fmt.Println(before) 
	
	arr := []string{before, response.PasswordChangeDate}
	sl := sort.StringSlice(arr)
	sl.Sort()
	fmt.Println(sl)

	if sl[0] == response.PasswordChangeDate{
		fmt.Println("update password") // updatePassword Login is in FindPassword/main.go
	}

	return c.JSON(http.StatusOK, response)
}


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

    fmt.Fprint(w, string(data)) // userinfo

	gUser := make(map[string]interface{})

   	if err := json.Unmarshal(data, &gUser); err != nil {
      fmt.Println(err.Error())
   }

   	fmt.Println(gUser)


	email = gUser["email"].(string)


	// 로그인
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := user{Id: "", LoginId: "", Password: "", Nickname: "", ProfileImg: "", PhoneNumber: "", LoginAttempt: "", PasswordChangeDate: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()
	
	var id string;
	var login_id string;
	var password string;
	var nickname string;
	var profile_img string;
	var phone_number string;
	var login_attempt string;
	var password_change_date string;

	err = db.QueryRow("SELECT * FROM user WHERE login_id = ?", email).Scan(&id, &login_id, &password, &nickname, &profile_img, &phone_number, &login_attempt, &password_change_date)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := user{Id: id, LoginId: login_id, Password: password, Nickname: nickname, ProfileImg: profile_img, PhoneNumber: phone_number, LoginAttempt: login_attempt, PasswordChangeDate: password_change_date}
	fmt.Println(response)
	fmt.Println(response.LoginId)

	if len(response.LoginId) == 0{
		return errors.New("가입되지 않은 아이디입니다")
	}

	fmt.Println("로그인 성공")
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

    return ioutil.ReadAll(res.Body)
}

func main() {

	// Echo instance
	e := echo.New()
  
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	  }))
	
	e.GET("/login", getUser) // login by local
	e.GET("/auth/google/login", getUserByGoogle) // login by google
   	e.GET("/auth/google/callback", googleAuthCallback)

	e.Logger.Fatal(e.Start(":3000"))
}