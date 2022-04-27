package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// loads all env vars from current dir
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	server := gin.Default()

	server.Run(":" + os.Getenv("PORT"))
}
