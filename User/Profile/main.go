package main

// 프로필 사진 업데이트
// http://localhost:3000//user/profile
// {
	// 	"login_id": "",
	// 	"profile_img": ""
// }

import (
	"database/sql"
	"fmt"
	"net/http"


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


func updateProfile(c echo.Context) error{
	// {
	// 	"login_id": "",
	// 	"profile_img": ""
	// }
	params := make(map[string]string)
    c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("UPDATE user set profile_img = ? WHERE login_id = ?", params["profile_img"],  params["login_id"])

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("update profile_img")
    }

    return c.JSON(http.StatusOK, result)
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
	
	e.PUT("/user/profile", updateProfile)

	e.Logger.Fatal(e.Start(":3000"))
}