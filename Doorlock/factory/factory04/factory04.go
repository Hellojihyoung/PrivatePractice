package factory04

import (
	"net/http"
	"fmt"


	"github.com/labstack/echo"
)

func PostUserCertificate(c echo.Context) error{
	doorlockId := c.Param("doorlockId")
	serialNumber := c.FormValue("serialNumber")
	CSR, err := c.FormFile("CSR")
	if err != nil {
		fmt.Print(err)
		return err
	}

	fmt.Println(doorlockId, serialNumber, CSR)

	return c.String(http.StatusOK, "")
}