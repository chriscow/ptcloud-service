package main

import (
	"strucim/gateway/services/identify"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
)

type dashboard struct {
	Title string
	Host  string
}

func main() {
	godotenv.Load()
	app := iris.New()

	app.RegisterView(iris.HTML("./templates", ".html")) // select the html engine to serve templates

	app.HandleDir("/js", "./static/js") // serve our custom javascript code.

	app.Get("/v1/locate/{service}", func(ctx iris.Context) {
		service := ctx.Params().Get("service")
		switch service {
		case "identify":
			ctx.JSON(iris.Map{
				"url":    "http://localhost:8080/v1/identify",
				"status": "ws://localhost:8080/v1/identify",
				"topic":  "identify.v1",
			})
		}
	})

	app.Post("/v1/identify", identify.PointCloud)

	websocketServer := websocket.New(websocket.DefaultGorillaUpgrader,
		identify.ServerEvents)
	

	app.Get("/v1/identify", websocket.Handler(websocketServer))

	app.Get("/", func(ctx iris.Context) {
		ctx.View("dashboard.html", dashboard{"Dashboard", "localhost:8080"})
	})

	app.Run(iris.Addr(":8080"))
}
