package doorlock14

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
)

type jwtAccessClaims struct {
	AuthType  int      `json:"authType"`
	UserId  int   `json:"userId"`
	MasterDoorlockIds []int `json:"masterDoorlockIds"`
	MemberDoorlockIds []int `json:"memberDoorlockIds"`
	IsTermAgree bool `json:"isTermAgree"`
	IsInfoRegistered bool `json:"isInfoRegistered"`
	IsPhoneVerified bool `json:"isPhoneVerified"`

	jwt.StandardClaims
}
type (
	tPassword struct {
		StartDate   string    `json:"startDate"`
		EndDate     string    `json:"endDate"`
		// DayList
		StartTime   string    `json:"startTime"`
		EndTime     string    `json:"endTime"`
		Nickname    string    `json:"nickname"`
		Password    string    `json:"password"`
		CreatedAt   string    `json:"createdAt"`
		CreateUser  string    `json:"createUser"`
		PasswordId  string    `json:"passwordId"`
	}
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

// Authorization
func verifyToken(r *http.Request) bool{
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	user := jwtAccessClaims{}
	token, _ := jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Valid)

	if !token.Valid{
		fmt.Println("invalid")
		return true 
	}

	return false
}

func GetTemporaryPassword(c echo.Context) error{
	
	// if verifyToken(c.Request()) {
	// 	return errors.New("authorization failed")
	// }

	doorlockId := c.Param("doorlockId")

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT start_date, end_date, start_time, end_time, nickname, password, created_at, make_user_id, id FROM door_lock_temporary_password WHERE door_lock_id = ?", doorlockId)

	if err != nil {
		fmt.Println(err)
	}
	
	defer rows.Close()

	// dayList 처리 요망
	var passwords = []tPassword{}

	for rows.Next() {
		var p tPassword
        err := rows.Scan(&p.StartDate, &p.EndDate, &p.StartTime, &p.EndTime, &p.Nickname, &p.Password, &p.CreatedAt, &p.CreateUser, &p.PasswordId)
        if err != nil {
            fmt.Println(err)
        }
		passwords = append(passwords, p)
    }
	
	return c.JSON(http.StatusOK, passwords)
}

func CreateTemporaryPassword(c echo.Context) error{
	
	// if verifyToken(c.Request()) {
	// 	return errors.New("authorization failed")
	// }

	params := make(map[string]string)
    c.Bind(&params)

	// daylist 처리하기
	doorlockId := c.Param("doorlockId")
	passwordType := params["passwordType"]
	userId := params["userId"]
	startDate := params["startDate"]
	endDate := params["endDate"]
	// dayList := params["dayList"]
	startTime := params["startTime"]
	endTime := params["endTime"]
	nickname := params["nickname"]
	password := params["password"]

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO invitation_code VALUES(0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, 0, 0, NOW())", doorlockId, userId, nickname, password, passwordType, startDate, endDate, startTime, endTime)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}

	
	return c.NoContent(http.StatusOK)
}

func DeleteTemporaryPassword(c echo.Context) error{
	
	// if verifyToken(c.Request()) {
	// 	return errors.New("authorization failed")
	// }

	doorlockId := c.Param("doorlockId")
	passwordId := c.Param("passwordId")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}
	
	
	result, err := db.Exec("DELETE FROM door_lock_temporary_password WHERE door_lock_id = ? AND id = ?", doorlockId, passwordId)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row deleted")
    }
	
	return c.NoContent(http.StatusOK)
}