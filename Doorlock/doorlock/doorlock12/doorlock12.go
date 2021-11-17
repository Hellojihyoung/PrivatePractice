package doorlock12

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

func UpdateDoorlockSetting(c echo.Context) error{

	// if verifyToken(c.Request()) {
	// 	return errors.New("Authorization failed")
	// }
	
	params := make(map[string]string)
	c.Bind(&params)

	doorlockId := c.Param("doorlockId")
	city := params["city"]
	district := params["district"]
	town := params["town"]


	db, err := sql.Open("mysql", connectionString)
    
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	result, err := db.Exec("UPDATE door_lock set city = ?, district = ?, town = ? WHERE id = ?", city, district, town, doorlockId)

	if err != nil {
    	fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

    if n == 1 {
    	fmt.Println("update")
   	}

   return c.NoContent(http.StatusOK)

}
