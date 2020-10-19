package routes

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"strucim/gateway/google"

	"github.com/kataras/iris/v12"
)

type identifyRequest struct {
	CsvFile  string `json:"file"`
	Callback string `json:"callback"`
}

// Identify is called when the client wants to identify a part using point
// cloud data. The file is sent embedded in a JSON wrapper as a CSV.
func Identify(ctx iris.Context) {

	req := &identifyRequest{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Failed to read JSON body").DetailErr(err))
		return
	}

	bucket, err := google.GetBucket(os.Getenv("IDENTIFY_BUCKET"))
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("GetBucket failed").DetailErr(err))
		return
	}

	csvBytes, err := base64.StdEncoding.DecodeString(req.CsvFile)
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Failed to decode file").DetailErr(err))
		return
	}

	csv := string(csvBytes)
	sr := strings.NewReader(csv)

	filename := fmt.Sprintf("pointcloud-%d", time.Now().Unix())
	if err := bucket.Store(filename, sr); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("bucket.Store() failed").DetailErr(err))
		return
	}

	if err := google.Publish(os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"), filename); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Publish() failed").DetailErr(err))
		return
	}

	// No error

}
