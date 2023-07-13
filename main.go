package main

import (
	"csi-api/app"
	"fmt"
	"os"
)

func main() {
	// if env.IsDevelopment() {
		app.Start("80")
		os.Setenv("MONGO_DB_NAME", "Test")
		fmt.Println(os.Getenv("MONGO_DB_NAME"))
	// } else {
	// 	app.Start(os.Getenv("PORT"))
	// }
}
