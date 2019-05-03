package main

import (
	"os"
	"log"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/afduarte/flickr-uploadr/handlers"
	"github.com/afduarte/flickr-uploadr/model"
	"github.com/joho/godotenv"
	"github.com/azer/go-flickr"
)

var (
	API_KEY    string
	API_SECRET string
	client     flickr.Client
)

func main() {
	API_KEY = os.Getenv("API_KEY")
	API_SECRET = os.Getenv("API_SECRET")
	if API_KEY == "" || API_SECRET == "" {
		godotenv.Load()
	}
	API_KEY = os.Getenv("API_KEY")
	API_SECRET = os.Getenv("API_SECRET")
	if API_KEY == "" || API_SECRET == "" {
		log.Fatal("Error reading Flickr API keys. Either load them as ENV variables or in a .env file")
	}
	// Initialise the model
	if dbErr := model.InitDB(); dbErr != nil {
		log.Fatal(dbErr)
	}
	defer model.DB.Close()
	// Create directory for files
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		os.Mkdir("files", 0755)
	}
	client = flickr.Client{
		Key: API_KEY,
	}
	go uploadManager()
	e := echo.New()
	e.Debug = true
	l := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	})
	e.Use(l)
	e.Use(middleware.Recover())

	e.Static("/", "static")
	e.POST("/upload", handlers.Upload)
	e.GET("/exists", handlers.CheckPhotoExists)
	e.POST("/job", handlers.AddJob)

	e.Logger.Fatal(e.Start(":1323"))
}

func uploadManager() {
	for {
		log.Println("Running cron")
		for _, job := range model.GetJobsDue() {
			go uploadr(job)
		}
		<-time.After(10 * time.Second)
	}
}

func uploadr(id []byte) {
	job := model.GetJob(id)
	log.Printf("Uploading Job: %v", *job)
	//filename := string(id)
}
