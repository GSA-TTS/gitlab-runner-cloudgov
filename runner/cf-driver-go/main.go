package main

import (
	"fmt"
	"log"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cf-driver/cmd"
	"github.com/joho/godotenv"
)

func main() {
	var err error

	defer func() {
		if err != nil {
			log.Println(fmt.Errorf("in main: %w", err))
		}

		if v := recover(); v != nil {
			log.Println(v)
		}
	}()

	if err := godotenv.Load(); err != nil {
		panic("error loading .env file")
	}

	cmd.Execute()
}
