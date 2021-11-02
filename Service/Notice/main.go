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
	id := c.FormValue("id")
	status := c.FormValue("status")
	title := c.FormValue("title")
	content := c.FormValue("content")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO NOTICE VALUES(?, ?, ?, ?)", id, status,title, content)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row inserted.")
    }

    return c.JSON(http.StatusOK, result)
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
	requested_page := c.QueryParam("page")
	requested_count := c.QueryParam("count")
	page , _ := strconv.Atoi(requested_page)
	count, _ := strconv.Atoi(requested_count)
	getPage := strconv.Itoa((page-1) * count)

	db, err := sql.Open("mysql", connectionString)
	
	if err != nil {
		fmt.Println(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, notice_title, notice_content FROM notice WHERE notice_status = 1 ORDER BY id DESC limit ?, ?", getPage, requested_count)

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


func deleteNotice(c echo.Context) error{
	requested_id := c.QueryParam("id")
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
	

	e.GET("/api/v1/service/notice", getAllNotices)
	e.POST("/v1/service/notice", createNotice)
	e.PUT("/v1/service/notice/:id", deleteNotice) // 삭제 _ status 변환

	e.GET("/notices/:id", getNotice)
	e.PUT("/notices/updateNotice/:id", updateNotice) // content 수정

	e.Logger.Fatal(e.Start(":3000"))
}