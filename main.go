package main

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/zinirun/eden-zip/src/zipper"
)

const fileName string = "edenzip-download.zip"
const staticDir string = "static/"

func splitUrls(v string) []string {
	return strings.Split(v, "\r\n")
}

func handleIndex(c echo.Context) error {
	return c.File(staticDir + "index.html")
}

func handleDownload(c echo.Context) error {
	urls := splitUrls(c.FormValue("urls"))
	zipper.Zipper(urls)
	return c.Attachment(fileName, fileName)
}

func main() {
	e := echo.New()
	e.GET("/", handleIndex)
	e.POST("/download", handleDownload)
	e.Logger.Fatal(e.Start(":1323"))
}
