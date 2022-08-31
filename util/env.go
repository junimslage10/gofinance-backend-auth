package util

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvParams struct {
	DbType               string `binding:"required"`
	DbDriver             string `binding:"required"`
	DbUser               string `binding:"required"`
	dbPassword           string `binding:"required"`
	DbHostDocker         string `binding:"required"`
	DbPort               string `binding:"required"`
	DbName               string `binding:"required"`
	DbSource             string `binding:"required"`
	ServerAddress        string `binding:"required"`
	FilledStringDbSource string `binding:"required"`
}

func LoadEnv() (DbDriver string, FilledStringDbSource string, ServerAddress string) {
	err := godotenv.Load("../.env")
	if err != nil {
		err = godotenv.Load()
	}
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	envArgs := &EnvParams{
		DbType:        os.Getenv("DB_TYPE"),
		DbDriver:      os.Getenv("DB_DRIVER"),
		DbUser:        os.Getenv("DB_USER"),
		dbPassword:    os.Getenv("DB_PASSWORD"),
		DbHostDocker:  os.Getenv("DB_HOST_DOCKER"),
		DbPort:        os.Getenv("DB_PORT"),
		DbName:        os.Getenv("DB_NAME"),
		DbSource:      os.Getenv("DB_SOURCE"),
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
	}

	var dbDriver = fmt.Sprintf(envArgs.DbDriver)
	var filledStringDbSource = fmt.Sprintf(
		envArgs.DbSource,
		envArgs.DbType,
		envArgs.DbUser,
		envArgs.dbPassword,
		envArgs.DbHostDocker,
		envArgs.DbPort,
		envArgs.DbName,
	)
	var serverAddress = fmt.Sprintf(envArgs.ServerAddress)

	return dbDriver, filledStringDbSource, serverAddress
}
