package doorlock

import (
	"database/sql"
	// "errors"
	"fmt"
	"log"
	"net/http"
	// "strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/golang-jwt/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)
const (
	host     = "localhost"
	database = "HDC"
	user1    = "root"
	password = "Ant123!!!"
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
	invitation struct {
		// InvitationId     int       `json:"invitationId"`
		InvitationCode   string    `json:"invitationCode"`
		// CreatedAt string `json:"createdAt"`
		ExpiresAt 		 string    `json:"expiresAt"`
		Nickname         string    `json:"nickname"`

	}
)
var (	
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user1, password, host, database)
)

//Doorlock-1
func CreateDoorlockId(c echo.Context) error{
	params := make(map[string]string)
    c.Bind(&params)

	serialNumber := c.QueryParam("serialNumber")
	city := params["city"]
	district := params["district"]
	town := params["town"]
	doorlockId := c.Param("doorlockId")


	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var id int; // 발급해준 값

	err = db.QueryRow("SELECT id FROM door_lock WHERE serial_number = ? AND city = ? AND dirstrict = ? AND town = ? AND id = ?", serialNumber, city, district, town, doorlockId).Scan(&id)

	if err != nil {
		fmt.Println(err)
	}
	
	return c.String(http.StatusOK, string(id))
}

//Doorlock-2
func DoorlockAuthenticate(c echo.Context) error{

	serialNumber := c.QueryParam("serialNumber")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	var id int; // 발급해준 값

	err = db.QueryRow("SELECT id FROM door_lock WHERE serial_number = ?", serialNumber).Scan(&id)
	
	if err != nil {
		fmt.Println(err)
	}
	
	return c.String(http.StatusOK, string(id))
}

// Authorization
func verifyToken(r *http.Request, userId int) bool{
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	user := jwtAccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Valid, user.UserId, err)

	if !token.Valid{
		fmt.Println("invalid")
		return true 
	}

	if user.UserId == userId{
		fmt.Print("-----------접근 가능 ------------")
		return false
	}
	return true
}

// Renewal Token
func renewalToken(r *http.Request) (map[string]string, error) {
	header := r.Header.Get("Authorization")
	tokenString := strings.Split(header, " ")[1]

	user := jwtAccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &user, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	log.Println(token.Valid, user.UserId, err)

	accessClaims := &jwtAccessClaims{
		user.AuthType,
		user.UserId,
		user.MasterDoorlockIds,
		user.MemberDoorlockIds,
		user.IsTermAgree,
		user.IsInfoRegistered,
		user.IsPhoneVerified,

		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte("secret"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	  return map[string]string{
		"access_token":  at,
	}, nil
}

//Doorlock-3
// func MapUserWithDoorlock(c echo.Context) error{
// 	params := make(map[string]string)
//     c.Bind(&params)

// 	// serialNumber := params["serialNumber"] : 
// 	userId, _ := strconv.Atoi(c.QueryParam("userId"))
// 	utype := c.QueryParam("type")
// 	doorlockId := c.Param("doorlockId")
// 	invitationCode := params["invitationCode"]

// 	if verifyToken(c.Request(), userId) {
// 		return errors.New("Authorization failed")
// 	}
	
// 	db, err := sql.Open("mysql", connectionString)

// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	defer db.Close()

// 	var authority int; // 발급해준 값

// 	err = db.QueryRow("SELECT authority FROM door_lock_user WHERE door_lock_id = ?", doorlockId).Scan(&authority)

// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	accessToken, _ := renewalToken(c.Request())
	
// 	return c.JSON(http.StatusOK, accessToken)
// }

// Doorlock-5
func GetInviation(c echo.Context) error{

	doorlockId := c.Param("doorlockId")

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	// Join 써서 낙네임을 빼와야할듯
	rows, err := db.Query("SELECT code, created_at, send_user_id FROM invitation_code WHERE door_lock_id = ?", doorlockId)

	if err != nil {
		fmt.Println(err)
	}
	
	defer rows.Close()

	var invitations = []invitation{}

	for rows.Next() {
		var i invitation
        err := rows.Scan(&i.InvitationCode, &i.ExpiresAt, &i.Nickname)
        if err != nil {
            fmt.Println(err)
        }
		invitations = append(invitations, i)
    }
	
	return c.JSON(http.StatusOK, invitations)
}


func CreateInviation(c echo.Context) error{
	// params := make(map[string]string)
	// c.Bind(&params)

	// doorlockId := c.Param("doorlockId")
	// nickname := params["nickname"]

	// db, err := sql.Open("mysql", connectionString)

	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// defer db.Close()

	// result, err := db.Exec("INSERT INTO STATE VALUES(?, ?, ?)", 0, serviceName, state)

	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// n, _ := result.RowsAffected()

	// if n == 1 {
	// 	fmt.Println("1 row inserted.")
	// }

    return c.NoContent(http.StatusOK)
}
