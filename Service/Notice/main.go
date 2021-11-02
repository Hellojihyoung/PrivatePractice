package main

// notice
import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	notice struct {
		Id   string    `json:"id"`
		Status   string    `json:"status"`
		Title string `json:"title"`
		Content string `json:"content"`
	}
)

const (
	host     = "localhost"
	database = "HDC"
	user     = "root"
	password = "Ant123!!!"
)

var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user, password, host, database)
	

func createNotice(c echo.Context) error{
	params := make(map[string]string)
    c.Bind(&params)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO NOTICE VALUES(?, ?, ?, ?)", params["id"], params["status"], params["title"], params["content"])

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row inserted.")
    }

    return c.JSON(http.StatusOK, params)
}

func getNotice(c echo.Context) error {
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := notice{Id: "", Status: "", Title: "", Content: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()
	
	var id string;
	var notice_title string;
	var notice_content string;

	err = db.QueryRow("SELECT id, notice_title, notice_content FROM notice WHERE id = ?", requested_id).Scan(&id, &notice_title, &notice_content)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := notice{Id: id, Title: notice_title, Content: notice_content}

	return c.JSON(http.StatusOK, response)
}

func getAllNotices(c echo.Context) error { // 보이는 것만 가져오기
	requested_page := c.Param("page")
	page , _ := strconv.Atoi(requested_page)
	getPage := strconv.Itoa((page-1) * 10)

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, notice_title, notice_content FROM notice WHERE notice_status = 1 limit "+ getPage +", 10")

	if err != nil {
		fmt.Println(err)
	}
	
	defer rows.Close()

	var notices = []notice{}

	for rows.Next() {
		var n notice
        err := rows.Scan(&n.Id, &n.Title, &n.Content)
        if err != nil {
            fmt.Println(err)
        }
		notices = append(notices, n)
    }

	return c.JSON(http.StatusOK, notices)
}

func updateNotice(c echo.Context) error{
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	params := make(map[string]string)
    c.Bind(&params)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("UPDATE NOTICE set notice_content = ? WHERE id = ?", params["content"], requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row updated.")
    }

    return c.JSON(http.StatusOK, result)
}


func updateVisability(c echo.Context) error{
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	params := make(map[string]string)
    c.Bind(&params)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("UPDATE NOTICE set notice_status = ? WHERE id = ?", params["status"], requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("modified visability")
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
	
	e.POST("/notices", createNotice)
	e.GET("/allNotices/:page", getAllNotices)
	e.GET("/notices/:id", getNotice)
	e.PUT("/notices/updateNotice/:id", updateNotice) // content 수정
	e.PUT("/notices/updateVisability/:id", updateVisability) // 삭제 _ status 변환
	// update를 한번에 하는 방법은..? 덮어씌워지지 않고 하는 방법

	e.Logger.Fatal(e.Start(":3000"))
}