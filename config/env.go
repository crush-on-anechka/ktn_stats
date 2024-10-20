package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	CredentialsFile string
	SheetParseRange string
	SpreadsheetIDs  []string
	SQLitePath      string
	TelegramToken   string
	TelegramChatID  int
	APIPort         int
}

var Envs = NewConfig()

func NewConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config := Config{
		CredentialsFile: getEnv("credentialsFile", ""),
		SheetParseRange: getEnv("sheetParseRange", SheetParseRange),
		SpreadsheetIDs:  getEnvAsSlice("spreadsheetIDString", ""),
		SQLitePath:      getEnv("SQLitePath", SQLitePath),
		TelegramToken:   getEnv("telegramToken", ""),
		TelegramChatID:  getEnvAsInt("telegramChatID", 0),
		APIPort:         getEnvAsInt("APIPort", 8000),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key, defaultValue string) []string {
	spreadsheetIDString := getEnv(key, defaultValue)
	valueAsSlice := strings.Split(spreadsheetIDString, ",")
	return valueAsSlice
}

func getEnvAsInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	valueAsInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueAsInt
}
