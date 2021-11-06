package main

import (
	"database/sql"
	"net/http"
	"fmt"

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
	login_id string
	profile_img string
)

func getUserInfo(c echo.Context) error{ 
	userId := c.Param("userId")
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

	e.GET("/api/v1/user/:userId", getUserInfo)

	e.Logger.Fatal(e.Start(":3000"))
}