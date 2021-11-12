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

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var id int;

	err = db.QueryRow("SELECT id FROM door_lock WHERE serial_number = ? AND city = ? AND dirstrict = ? AND town = ?", serialNumber, city, district, town).Scan(&id)

	if err != nil {
		fmt.Println(err)
	}
	
	return c.String(http.StatusOK, string(id))
}
