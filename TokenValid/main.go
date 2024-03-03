package main

import (
	"TokenValid/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()

	router.POST("/token/generate", handlers.GenerateToken)
	router.POST("/token/validate", handlers.ValidateToken)

	router.Run(":8082")
}
