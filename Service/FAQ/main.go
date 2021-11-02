package main

// faq
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
	faq struct {
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
	

func createFAQ(c echo.Context) error{
	id := c.FormValue("id")
	status := c.FormValue("status")
	title := c.FormValue("title")
	content := c.FormValue("content")

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("INSERT INTO FAQ VALUES(?, ?, ?, ?)", id, status, title, content)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row inserted.")
    }

    return c.JSON(http.StatusOK, result)
}

func getFAQ(c echo.Context) error {
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		response := faq{Id: "", Status: "", Title: "", Content: ""}
			return c.JSON(http.StatusInternalServerError, response)
	}
	defer db.Close()
	
	var id string;
	var faq_title string;
	var faq_content string;

	err = db.QueryRow("SELECT id, notice_title, notice_content FROM notice WHERE id = ?", requested_id).Scan(&id, &faq_title, &faq_content)
	
	if err != nil {
		fmt.Println(err)
	}
	
	response := faq{Id: id, Title: faq_title, Content: faq_content}

	return c.JSON(http.StatusOK, response)
}

func getAllFAQs(c echo.Context) error { // 보이는 것만 가져오기

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

	rows, err := db.Query("SELECT id, faq_title, faq_content FROM faq WHERE faq_status = 1 ORDER BY id DESC limit ?, ?", getPage, requested_count)

	if err != nil {
		fmt.Println(err)
	}
	
	defer rows.Close()

	var faqs = []faq{}

	for rows.Next() {
		var f faq
        err := rows.Scan(&f.Id, &f.Title, &f.Content)
        if err != nil {
            fmt.Println(err)
        }
		faqs = append(faqs, f)
    }

	return c.JSON(http.StatusOK, faqs)
}

func updateFAQContent(c echo.Context) error{
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	params := make(map[string]string)
    c.Bind(&params)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("UPDATE FAQ set FAQ_content = ? WHERE id = ?", params["content"], requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row content updated.")
    }

    return c.JSON(http.StatusOK, result)
}

func updateFAQTitle(c echo.Context) error{
	requested_id := c.Param("id")
	db, err := sql.Open("mysql", connectionString)

	params := make(map[string]string)
    c.Bind(&params)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("UPDATE FAQ set FAQ_content = ? WHERE id = ?", params["content"], requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row content updated.")
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

	result, err := db.Exec("UPDATE FAQ set faq_status = ? WHERE id = ?", params["status"], requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("modified visability")
    }

    return c.JSON(http.StatusOK, result)
}


func deleteFAQ(c echo.Context) error{
	requested_id := c.QueryParam("id")
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer db.Close()

	result, err := db.Exec("DELETE FROM FAQ WHERE id = ?", requested_id)

	if err != nil {
        fmt.Println(err.Error())
    }

    n, _ := result.RowsAffected()

    if n == 1 {
        fmt.Println("1 row deleted")
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
	
	e.GET("/api/v1/service/faq", getAllFAQs)
	e.POST("/v1/service/faq", createFAQ)
	e.DELETE("/v1/service/faq", deleteFAQ)


	
	e.GET("/faqs/:id", getFAQ)
	e.PUT("/faqs/update/content/:id", updateFAQContent) // content 수정
	e.PUT("/faqs/update/title/:id", updateFAQTitle) // title 수정
	e.PUT("/faqs/updateVisability/:id", updateVisability) // 삭제 _ status 변환
	// update를 한번에 하는 방법은..? 덮어씌워지지 않고 하는 방법

	e.Logger.Fatal(e.Start(":3000"))
}