package doorlock02

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
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

func AuthentificateDoolock(c echo.Context) error{
	params := make(map[string]string)
	c.Bind(&params)

	serialNumber := params["serialNumber"]
	city := params["city"]
	district := params["district"]
	town := params["town"]

	db, err := sql.Open("mysql", connectionString)
    
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}

	result, err := db.Exec("UPDATE door_lock set city = ?, district = ?, town = ? WHERE serial_number = ?", city, district, town, serialNumber)

	if err != nil {
    	fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

    if n == 1 {
    	fmt.Println("update location")
   	}
 
	var id string;

	_ = db.QueryRow("SELECT id FROM door_lock WHERE serial_number = ?", serialNumber).Scan(&id)

	return c.String(http.StatusOK, id)
}
