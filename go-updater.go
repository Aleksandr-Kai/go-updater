package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	domain       = "https://go.dev"
	downloadPage = "https://go.dev/dl"
	installPath  = "/usr/local/go"
	downloadPath = "/Downloads/golang.tar.gz"
)

func request(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getDownloadURL(r io.ReadCloser) (string, error) {
	defer r.Close()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}
	a := doc.Find(`.downloadBox[href*="linux"]`).First()
	if a == nil {
		return "", errors.New("download link not found")
	}
	href, ok := a.Attr("href")
	if !ok {
		return "", errors.New("download link not found")
	}
	return "https://go.dev" + href, nil
}

func saveFile(file io.ReadCloser) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	out, err := os.Create(home + downloadPath)
	if err != nil {
		return err
	}
	defer out.Close()
	defer file.Close()
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(file, counter)); err != nil {
		return err
	}
	fmt.Println("\nDone")
	return nil
}

func removeFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.Remove(home + downloadPath)
}

func install() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cmd := `sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ` + home + downloadPath
	out, err := exec.Command("bash", "-c", cmd).Output()
	fmt.Println(string(out))
	if err != nil {
		return err
	}
	return nil
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func main() {
	logrus.Info("Get download URL")
	resp, err := request(downloadPage)
	if err != nil {
		logrus.Error(err)
		return
	}

	dl, err := getDownloadURL(resp)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("Download file: ", dl)
	file, err := request(dl)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer file.Close()
	if err = saveFile(file); err != nil {
		logrus.Error(err)
		return
	}
	defer removeFile()
	logrus.Info("Install...")
	if err = install(); err != nil {
		logrus.Error(err)
	}
	out, err := exec.Command("bash", "-c", "go version").Output()
	fmt.Println(string(out))
	if err != nil {
		logrus.Error(err)
	}
}
