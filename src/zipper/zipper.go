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
	randomPath := "tmp/" + uriuri.New() + "/"
	filenames := []string{}
	chForDownload := make(chan string)

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

	if len(filenames) == 0 {
		return "", "There's no files to download avaliable."
	}

	fmt.Println("Start Making zip ...")

	files := make(chan *os.File)
	zipWait := writeZip(zipfilename, files, randomPath)
	var wg sync.WaitGroup
	wg.Add(len(filenames))
	for _, filename := range filenames {
		go func(name string) {
			defer wg.Done()
			f, err := os.Open(name)
			if err != nil {
				panic(err)
			}
			files <- f
		}(randomPath + filename)
	}

	wg.Wait()
	close(files)
	zipWait.Wait()

	defer func() {
		for _, f := range filenames {
			os.Remove(randomPath + f)
		}
	}()

	return randomPath, ""
}

func writeZip(outFilename string, files chan *os.File, path string) *sync.WaitGroup {
	outf, err := os.Create(path + outFilename)
	errorHandler(err)
	var wg sync.WaitGroup
	wg.Add(1)
	zw := zip.NewWriter(outf)
	go func() {
		defer wg.Done()
		defer outf.Close()
		var err error
		var fw io.Writer
		for f := range files {
			if fw, err = zw.Create(filepath.Base(f.Name())); err != nil {
				panic(err)
			}
			io.Copy(fw, f)
			if err = f.Close(); err != nil {
				panic(err)
			}
		}
		if err = zw.Close(); err != nil {
			panic(err)
		}
	}()
	return &wg
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
