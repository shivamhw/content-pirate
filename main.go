package main

import (
	"github.com/shivamhw/content-pirate/cmd"
	"github.com/shivamhw/content-pirate/pkg/log"
		"github.com/joho/godotenv"

)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	cmd.Execute()
}
