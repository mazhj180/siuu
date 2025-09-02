package main

import (
	"siuu/internal/clicmd"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	_ = clicmd.Execute()
}
