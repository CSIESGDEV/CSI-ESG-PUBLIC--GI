package main

import (
	"csi-api/app"
	"csi-api/app/env"
	"fmt"
	"os"
)

func main() {
	if env.IsDevelopment() {
		app.Start("5000")
		os.Setenv("MONGO_DB_NAME", "Test")
		fmt.Println(os.Getenv("MONGO_DB_NAME"))
	} else {
		app.Start(os.Getenv("PORT"))
	}
}
