package doorlock16

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
	model struct {
		Version   string    `json:"version"`
		Expires     string    `json:"expires"`
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

func GetModel(c echo.Context) error{
	
	// if verifyToken(c.Request()) {
	// 	return errors.New("authorization failed")
	// }

	modelNumber := c.Param("modelNumber")

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var version string;
	var expires string;

	err = db.QueryRow("SELECT version, release_time FROM firmware WHERE id = ?", modelNumber).Scan(&version, &expires)

	if err != nil {
		fmt.Println(err)
	}

	response := model{Version: version, Expires: expires}
	
	return c.JSON(http.StatusOK, response)
}