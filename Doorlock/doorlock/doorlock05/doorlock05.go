package doorlock05

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

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
	invitation struct {
		// InvitationId     int       `json:"invitationId"`
		InvitationCode   string    `json:"invitationCode"`
		CreatedAt        string    `json:"createdAt"`
		// ExpiresAt 		 string    `json:"expiresAt"`
		Nickname         string    `json:"nickname"`

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

func GetInviation(c echo.Context) error{
	
	if verifyToken(c.Request()) {
		return errors.New("authorization failed")
	}

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
        err := rows.Scan(&i.InvitationCode, &i.CreatedAt, &i.Nickname)
        if err != nil {
            fmt.Println(err)
        }
		invitations = append(invitations, i)
    }
	
	return c.JSON(http.StatusOK, invitations)
}


func receiveId(nickname string) int{
	
	db, err := sql.Open("mysql", connectionString)
    
	defer db.Close()

	if err != nil {
		fmt.Println(err)
	}
 
	var receiveUserId int;

	_ = db.QueryRow("SELECT user_id FROM door_lock_user WHERE nickname = ?", nickname).Scan(&receiveUserId)

	return receiveUserId
}

func generateInvitationCode() string { // 초대코드
	var randNum = []rune("0123456789")

	s := make([]rune, 6)
	rand.Seed(time.Now().UnixNano())
	time.Sleep(10)
	for i := range s {
		s[i] = randNum[rand.Intn(len(randNum))]
	}
	return string(s)
}

func CreateInviation(c echo.Context) error{
	
	if verifyToken(c.Request()) {
		return errors.New("authorization failed")
	}

	params := make(map[string]string)
    c.Bind(&params)

	doorlockId := c.Param("doorlockId")
	nickname := params["nickname"]
	receiveUserId := receiveId(nickname)
	code := generateInvitationCode()

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO invitation_code VALUES(?, ?, ?, ?, 0, NOW())", receiveUserId, receiveUserId, doorlockId, code)

	if err != nil {
		fmt.Println(err.Error())
	}

	n, _ := result.RowsAffected()

	if n == 1 {
		fmt.Println("1 row inserted.")
	}

	
	return c.NoContent(http.StatusOK)
}

func DeleteInviation(c echo.Context) error{
	
	if verifyToken(c.Request()) {
		return errors.New("authorization failed")
	}

	doorlockId := c.Param("doorlockId")
	invitationId := c.Param("invitionId")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}
	
	
	result, err := db.Exec("DELETE FROM invitation_code WHERE door_lock_id = ? AND code = ?", doorlockId, invitationId)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row deleted")
    }
	
	return c.NoContent(http.StatusOK)
}