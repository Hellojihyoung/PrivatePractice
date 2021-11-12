package main

import (
	
	"github.com/labstack/echo"
	"doorlock/factory"
	"doorlock/factory/factory02"
	"doorlock/factory/factory04"
	// "doorlock/doorlock"
	"doorlock/doorlock/doorlock02"
	"doorlock/doorlock/doorlock03"

)


func main() {
	e := echo.New()
	d := e.Group("/api/v1/doorlocks")
	u := e.Group("/api/v1/users")
	
	d.POST("/serial-number", factory.CreateSerialNumber) // [Factory-1]
	d.GET("/factory-certificate", factory02.GetFactoryCertificate) // [Factory-2]
	d.POST("/:doorlockId/issue-user-certificate", factory04.PostUserCertificate) // [Factory-4]

	// d.POST("/check-serial-number", doorlock.CheckSerialNumber) // [DOORLOCK-1]
	d.POST("/authentificate", doorlock02.AuthentificateDoolock) // [DOORLOCK-2]
	u.POST("/:userId/bind-doorlock", doorlock03.MapUserWithDoorlock) // [DOORLOCK-3]




	// d.POST("/check-sn", doorlock.CreateDoorlockId)
	// d.POST("/:doorlockId/authenticate", doorlock.DoorlockAuthenticate)
	// d.GET("/:doorlockId/invitation", doorlock.GetInviation) // 5
	// d.POST("/:doorlockId/invitation", doorlock.CreateInviation) // 5
	// d.DELETE("/:doorlockId/invitation", doorlock.DeleteInviation) // 5


	



	e.Logger.Fatal(e.Start(":3000"))
}

