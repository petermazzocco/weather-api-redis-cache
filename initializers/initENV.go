package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

func InitENV() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or cannot be loaded - using environment variables instead")
	}
}
