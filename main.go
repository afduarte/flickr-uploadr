package main

import (
	"os"
	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/afduarte/flickr-uploadr/handlers"
	"github.com/afduarte/flickr-uploadr/DB"
)

func main() {
	dbErr := DB.Init()
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	defer DB.DB.Close()
	// Create directory for files
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		os.Mkdir("files", 0755)
	}
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "static")
	e.POST("/upload", handlers.Upload)
	e.GET("/exists", handlers.CheckPhotoExists)

	e.Logger.Fatal(e.Start(":1323"))
}
