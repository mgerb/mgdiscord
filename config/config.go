package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config - global config object
var Config = &configType{}

type configType struct {
	Token     string
	BotPrefix string
}

// Init -
func Init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	Config.Token = os.Getenv("TOKEN")
	Config.BotPrefix = os.Getenv("BOT_PREFIX")

	if Config.Token == "" {
		log.Fatal("Please specify environment variable: TOKEN")
	}

	if Config.BotPrefix == "" {
		log.Println("BOT_PREFIX not set - defaulting to !")
		Config.BotPrefix = "!"
	}
}
