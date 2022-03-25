package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getEnv(key string) string {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("É necessário informar a variável de ambiente: %s", key))
	}

	return value
}
