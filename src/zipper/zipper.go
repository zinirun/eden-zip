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
	"sync"

	"github.com/zinirun/eden-zip/src/uriuri"
)

var (
	errBadStatus   error = errors.New("Bad Status Exists while http.Get")
	errGetFilePath error = errors.New("Can't get filepath")
	errOSCreate    error = errors.New("Can't os.Create")
	errIOCopy      error = errors.New("Can't io.Copy")
)

// Zipper make .zip file from urls
func Zipper(urls []string, zipfilename string) (string, string) {
	filenames := []string{}
	chForDownload := make(chan string)

	randomPath := "tmp/" + uriuri.New() + "/"
	os.Mkdir(randomPath, 0755)

	fmt.Println("Start Downloading ...")

	for _, url := range urls {
		go func(url string, c chan<- string) {
			f, err := downloadFile(url, randomPath)
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
		return "", "There's no files to download avaliable."
	}
	err := writeZip(zipfilename, filenames, randomPath)
	if err != nil {
		return "", "Errors occured while making zip."
	}
	return randomPath, ""
}

func writeZip(outFilename string, filenames []string, path string) error {
	outf, err := os.Create(path + outFilename)
	errorHandler(err)
	var wg sync.WaitGroup
	zw := zip.NewWriter(outf)
	for _, filename := range filenames {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			w, err := zw.Create(filename)
			errorHandler(err)
			f, err := os.Open(path + filename)
			errorHandler(err)
			defer f.Close()
			_, err = io.Copy(w, f)
			errorHandler(err)
		}(filename)
	}
	wg.Wait()
	defer func() {
		for _, filename := range filenames {
			os.Remove(path + filename)
		}
	}()
	return zw.Close()
}

func downloadFile(url string, path string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errBadStatus
	}
	filename, err := getFileName(url)
	if err != nil {
		return "", errGetFilePath
	}
	f, err := os.Create(path + filename)
	fmt.Println(f)
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
