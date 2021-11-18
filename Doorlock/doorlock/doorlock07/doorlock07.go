package doorlock07

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
	doorlock struct {
		ModelNumber     string  `json:"modelNumber"`
		// Nickname        string  `json:"nickname"`
		Battery         int     `json:"battery"`
		Cell            int     `json:"cell"`
		HomeSafe        bool    `json:"homeSafe"`
		Firmware        string  `json:"firmware"`
		SerialNumber    string  `json:"serialNumber"`
		City            string  `json:"city"`
		District        string  `json:"district"`
		Town            string  `json:"town"`
		CertCode        string  `json:"certCode"`

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

func GetDoorlockInfo(c echo.Context) error{
	// if verifyToken(c.Request()) {
	// 	return errors.New("Authorization failed")
	// }

	doorlockId := c.Param("doorlockId")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		response := doorlock{ModelNumber: ""}
		return c.JSON(http.StatusInternalServerError, response)
	}

	defer db.Close()

	var modelNumber string;
	// var nickname string; // 다른 곳에서 가져오기
	var battery int;
	var cell int;
	// users
	var homeSafe bool;
	// wifi
	var firmware string;
	var serialNumber string;
	var city string;
	var district string;
	var town string;
	var certCode string;

	err = db.QueryRow("SELECT model_name, battery, cell, is_home_safe, firmware_id, serial_number, city, district, town, auth_code FROM door_lock WHERE id = ?", doorlockId).Scan(&modelNumber, &battery, &cell, &homeSafe, &firmware, &serialNumber, &city, &district, &town, &certCode)
	
	if err != nil {
		fmt.Println(err)
	}

	response := doorlock{ModelNumber: modelNumber, Battery: battery, Cell: cell, HomeSafe: homeSafe, Firmware: firmware, SerialNumber: serialNumber, City: city, District: district , Town: town, CertCode : certCode}

	return c.JSON(http.StatusOK, response)
}


func UpdateDoorlockSetting(c echo.Context) error{

	// if verifyToken(c.Request()) {
	// 	return errors.New("Authorization failed")
	// }
	
	params := make(map[string]string)
	c.Bind(&params)

	doorlockId := c.Param("doorlockId")
	name := params["name"]

	db, err := sql.Open("mysql", connectionString)
    
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	result, err := db.Exec("UPDATE door_lock set name = ? WHERE id = ?", name, doorlockId)

	if err != nil {
    	fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

    if n == 1 {
    	fmt.Println("update")
   	}

   return c.NoContent(http.StatusOK)

}
