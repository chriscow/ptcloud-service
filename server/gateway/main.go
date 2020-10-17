package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

func main() {
	// load project-defined environment variables from .env file
	godotenv.Load()

	// iris is a web framework similar to Flask / Django
	app := iris.Default() // use the default config

	app.Get("/v1/heartbeat", func(ctx iris.Context) {
		// write out a small JSON response
		ctx.JSON(iris.Map{
			"status":    "OK",
			"timestamp": time.Now().UTC().Unix(),
		})
	})

	app.Post("/v1/identify", func(ctx iris.Context) {
		// expecting the part point cloud data in the body
		// write it to cloud storage, then send a pub/sub message
		// with its location
		file, _, err := ctx.FormFile("file")
		if err != nil {
			app.Logger().Error("Could not read posted file")
			ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
				Title("reading form file").DetailErr(err))
			return
		}

		bucket, err := GetBucket(os.Getenv("IDENTIFY_BUCKET"))
		if err != nil {
			app.Logger().Errorf("GetBucket failed: %v", err)
			ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
				Title("GetBucket").DetailErr(err))
			return
		}

		filename := fmt.Sprintf("pointcloud-%d", rand.Int())
		if err := bucket.Store(filename, file); err != nil {
			app.Logger().Errorf("bucket.Store() failed: %v", err)
			ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
				Title("bucket.Store").DetailErr(err))
			return
		}

		if err := Publish(os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"), filename); err != nil {
			app.Logger().Errorf("Publish() failed: %v", err)
			ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
				Title("Publish").DetailErr(err))
			return
		}

		// No error
		// ctx.StopWithText(iris.StatusAccepted, "OK")
		ctx.JSON(iris.Map{
			"status": "OK",
		})

		// app.Logger().Debugf("tada", len(bytes))
	})

	app.Post("/v1/session", func(ctx iris.Context) {
		// expecting session metadata in JSON format
		// session := &models.Session{}
		// if err := ctx.ReadJSON(&session); err != nil {
		// 	ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
		// 		Title("Session creation failure").DetailErr(err))

		// 	return
		// }

		// // generate a unique session id
		// // create a bucket containing the session id in its name
		// // return the session json
		// app.Logger().Debugf("%#v", session)

		// session.ID = rand.Int63()

		// // create a bucket and fill in the Bucket struct with the info
		// bucket := &models.Bucket{Name: iota(session.ID)}
		// bucket.Create()
		// ctx.JSON(bucket)
	})

	app.Listen(":8080")
}
