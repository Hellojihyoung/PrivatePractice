package main

import (
	
	"github.com/labstack/echo"
	"doorlock/factory"
	"doorlock/doorlock"
)

func main() {
	e := echo.New()
	d := e.Group("/api/v1/doorlocks")
	
	d.POST("/serial-number", factory.CreateSerialNumber) // [Factory-1]


	d.POST("/check-sn", doorlock.CreateDoorlockId)
	d.POST("/:doorlockId/authenticate", doorlock.DoorlockAuthenticate)

	// d.POST("/:doorlockId/bind-doorlock", doorlock.MapUserWithDoorlock) // 3
	d.GET("/:doorlockId/invitation", doorlock.GetInviation) // 5
	d.POST("/:doorlockId/invitation", doorlock.CreateInviation) // 5
	// d.DELETE("/:doorlockId/invitation", doorlock.DeleteInviation) // 5



	e.Logger.Fatal(e.Start(":3000"))
}

