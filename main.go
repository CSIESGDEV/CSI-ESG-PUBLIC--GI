package main

import (
	"os"
	"sme-api/app"
	"sme-api/app/env"
)

func main() {
	if env.IsDevelopment() {
		app.Start("5000")
	} else {
		app.Start(os.Getenv("PORT"))
	}
}
