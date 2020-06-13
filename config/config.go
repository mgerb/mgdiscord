package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"

	"github.com/joho/godotenv"
)

// Config - global config object
var Config = &configType{}

type configType struct {
	Token     string `required:"true"`
	BotPrefix string `default:"!" split_words:"true"`
	Timeout   int    `default:"30"`
}

// Init -
func Init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found - using environment variables")
	}

	err = envconfig.Process("", Config)

	if err != nil {
		log.Fatal(err)
	}
}
