package main

import (
	"skinSync/config"
	"skinSync/middlewares"
	"skinSync/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.ConnectDB()
	e := echo.New()
	e.Binder = &middlewares.CustomBinder{}

	// Middleware for panic recover
	e.Use(middleware.Recover())
	router.RequestHandler(e)
}
