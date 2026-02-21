package main

import (
	"skinSync/config"
	"skinSync/middlewares"
	"skinSync/router"
	"skinSync/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.ConnectDB()

	// Start background cleanup goroutines
	services.StartOTPCleanup()
	services.StartTokenBlacklistCleanup()

	e := echo.New()
	e.Binder = &middlewares.CustomBinder{}

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middlewares.CORSMiddleware())
	router.RequestHandler(e)
}
