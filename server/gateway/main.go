package main

import (
	"strucim/gateway/routes"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

func main() {
	// load project-defined environment variables from .env file
	godotenv.Load()

	// iris is a web framework similar to Flask / Django
	app := iris.Default() // use the default config

	app.Get("/v1/heartbeat", routes.Heartbeat)

	app.Post("/v1/identify", routes.Identify)

	app.Listen(":8080")
}
