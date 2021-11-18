package doorlock10

import (
	"database/sql"
	// "errors"
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
	wifi struct {
		WifiPassword    string  `json:"wifiPassword"`
		WifiStrength    string  `json:"wifiStrength"`
		WifiName        string  `json:"wifiName"`
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

func GetDoorlockWifi(c echo.Context) error{
	// if verifyToken(c.Request()) {
	// 	return errors.New("Authorization failed")
	// }

	doorlockId := c.Param("doorlockId")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error")
	}

	defer db.Close()

	var wifiPassword string;
	var wifiStrength string;
	var wifiName string;

	err = db.QueryRow("SELECT wifi_password, wifi_strength, wifi_name FROM door_lock WHERE id = ?", doorlockId).Scan(&wifiPassword, &wifiStrength, &wifiName)
	
	if err != nil {
		fmt.Println(err)
	}

	response := wifi{WifiPassword: wifiPassword, WifiStrength: wifiPassword, WifiName: wifiName}


	return c.JSON(http.StatusOK, response)
}