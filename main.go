package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/zinirun/eden-zip/src/zipper"
)

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
	filePath, errString := zipper.Zipper(urls, filename)
	if errString != "" {
		return c.String(http.StatusOK, "Sorry, error occured: "+errString)
	}
	return c.Attachment(filePath+filename, filename)
}

func main() {
	defer os.Remove("tmp")
	os.Mkdir("tmp", 0755)
	e := echo.New()
	e.GET("/", handleIndex)
	e.POST("/download", handleDownload)
	e.Logger.Fatal(e.Start(":1323"))
}
