package zipper

/*

https://github.com/zinirun/zinirun/raw/main/icons/go.png
https://github.com/zinirun/zinirun/raw/main/icons/typescript.png
https://github.com/zinirun/zinirun/raw/main/icons/nodejs.png
https://github.com/zinirun/zinirun/raw/main/icons/nestjs.png

*/

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var (
	errBadStatus   error = errors.New("Bad Status Exists while http.Get")
	errGetFilePath error = errors.New("Can't get filepath")
	errOSCreate    error = errors.New("Can't os.Create")
	errIOCopy      error = errors.New("Can't io.Copy")
)

// Zipper make .zip file from urls
func Zipper(urls []string, zipfilename string) string {
	filenames := []string{}
	chForDownload := make(chan string)

	fmt.Println("Start Downloading ...")

	for _, url := range urls {
		go func(url string, c chan<- string) {
			f, err := downloadFile(url)
			if err != nil {
				fmt.Println(err)
				c <- "ERR"
				return
			}
			c <- f
		}(url, chForDownload)
	}

	for range urls {
		go func(filename string) {
			if filename == "ERR" {
				return
			}
			filenames = append(filenames, filename)
		}(<-chForDownload)
	}

	fmt.Println("Start Making zip ...")

	if len(filenames) == 0 {
		return "There's no files to download avaliable."
	}
	err := writeZip(zipfilename, filenames)
	if err != nil {
		return "Errors occured while making zip."
	}
	return ""
}

func writeZip(outFilename string, filenames []string) error {
	c := make(chan bool)
	outf, err := os.Create(outFilename)
	errorHandler(err)
	zw := zip.NewWriter(outf)
	for _, filename := range filenames {
		go func(filename string, c chan<- bool) {
			w, err := zw.Create(filename)
			errorHandler(err)
			f, err := os.Open(filename)
			errorHandler(err)
			defer f.Close()
			_, err = io.Copy(w, f)
			errorHandler(err)
			c <- true
		}(filename, c)
	}
	for range filenames {
		<-c
	}
	defer func() {
		for _, filename := range filenames {
			os.Remove(filename)
		}
	}()
	return zw.Close()
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errBadStatus
	}
	filename, err := getFileName(url)
	if err != nil {
		return "", errGetFilePath
	}
	f, err := os.Create(filename)
	if err != nil {
		return "", errOSCreate
	}
	defer func() {
		resp.Body.Close()
		f.Close()
	}()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", errIOCopy
	}
	return filename, nil
}

func getFileName(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", errGetFilePath
	}
	return filepath.Base(url.Path), nil
}

func errorHandler(err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
