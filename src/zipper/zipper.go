package zipper

/*

func main() {
	urls := []string{
		"https://github.com/zinirun/zinirun/raw/main/icons/go.png",
		"https://github.com/zinirun/zinirun/raw/main/icons/typescript.png",
		"https://github.com/zinirun/zinirun/raw/main/icons/nodejs.png",
		"https://github.com/zinirun/zinirun/raw/main/icons/nestjs.png",
	}
	synczip.AdvZipper(urls)
}

*/

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Zipper make .zip file from urls
func Zipper(urls []string) {
	filenames := []string{}
	chForDownload := make(chan string)

	fmt.Println("Start Downloading ...")

	for _, url := range urls {
		go func(url string, c chan<- string) {
			f, err := downloadFile(url)
			errorLogHandler(err)
			c <- f
		}(url, chForDownload)
	}

	for range urls {
		filenames = append(filenames, <-chForDownload)
	}

	fmt.Println("Start Making zip ...")

	err := writeZip("edenzip-download.zip", filenames)
	errorLogHandler(err)
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
	errorStringHandler(err)
	filename, err := getFileName(url)
	errorStringHandler(err)
	f, err := os.Create(filename)
	errorStringHandler(err)
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return filename, err
}

func getFileName(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return filepath.Base(url.Path), nil
}

func errorLogHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errorHandler(err error) error {
	if err != nil {
		return err
	}
	return nil
}

func errorStringHandler(err error) (string, error) {
	if err != nil {
		return "", err
	}
	return "", nil
}
