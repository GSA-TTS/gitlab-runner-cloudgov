package main

import (
	"fmt"
	"log"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cf-executor/cg"
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

	cgClient, err := cg.New(&cg.GoCFClientAdapter{}, nil)
	if err != nil {
		panic(err)
	}

	apps, err := cgClient.GetApps()
	if err != nil {
		panic(err)
	}

	for _, app := range apps {
		log.Printf("Application %s is %s\n", app.Name, app.State)
	}
}
