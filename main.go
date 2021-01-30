package main

import (
	"net/http"
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
	filename := c.FormValue("filename")
	if filename == "" {
		filename = "edenzip-download.zip"
	}
	errString := zipper.Zipper(urls, filename)
	if errString != "" {
		return c.String(http.StatusOK, "Sorry, error occured: "+errString)
	}
	return c.Attachment(fileName, fileName)
}

func main() {
	e := echo.New()
	e.GET("/", handleIndex)
	e.POST("/download", handleDownload)
	e.Logger.Fatal(e.Start(":1323"))
}
