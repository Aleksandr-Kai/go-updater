package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
)

var (
	downloadPage = "https://go.dev/dl"
	installPath  = "/usr/local"
	downloadPath = "~/Downloads/golang.tar.gz"
	domain       = regexp.MustCompile(`^\w+://[\w.-]+`)
	rgoversion   = regexp.MustCompile(`go[0-9]+[0-9\.]+[0-9]+`)
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

func getDownloadURL() (string, error) {
	r, err := request(downloadPage)
	if err != nil {
		return "", err
	}

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
	return domain.FindString(downloadPage) + href, nil
}

func saveFile(file io.ReadCloser, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	defer file.Close()
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(file, counter))
	return err
}

func AskUserYN(message string, defYes bool) string {
	reader := bufio.NewReader(os.Stdin)
	answer := "no answer"
	for answer != "y" && answer != "Y" && answer != "n" && answer != "N" && answer != "" {
		if defYes {
			fmt.Printf("%s (Y/n): ", message)
		} else {
			fmt.Printf("%s (y/N): ", message)
		}
		answer, _ = reader.ReadString('\n')
		answer = strings.Trim(answer, "\n")
	}

	return strings.ToLower(answer)
}

func downloadGo(url string) error {
	if _, err := os.Stat(downloadPath); !errors.Is(err, os.ErrNotExist) {
		if AskUserYN("File exists. Replace it?", true) == "n" {
			return nil
		}
	}
	fmt.Println("Download file: ", url)
	file, err := request(url)
	if err != nil {
		return err
	}
	defer file.Close()
	defer fmt.Println("")
	return saveFile(file, downloadPath)
}

func install(destPath, sourcePath string) error {
	cmdString := spew.Sprintf("sudo rm -rf %s/go && sudo tar -C %s -xzf %s", destPath, destPath, sourcePath)
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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

func getGoVersionInstalled() (string, error) {
	out, err := exec.Command("bash", "-c", "go version").Output()
	ver := rgoversion.Find(out)
	fmt.Printf("Installed version: %s\n", string(ver))
	return string(ver), err
}

func getGoVersionActual(url string) string {
	ver := rgoversion.FindString(url)
	fmt.Printf("Current version: %s\n", ver)
	return string(ver)
}

func getInstallPath() (string, error) {
	out, err := exec.Command("bash", "-c", "which go").Output()
	if err != nil {
		return "", fmt.Errorf("getInstallPath: unable to get path to 'go'")
	}

	r := regexp.MustCompile("/go.+")
	indexes := r.FindIndex(out)
	if len(indexes) == 0 {
		return "", fmt.Errorf("getInstallPath: invalid path")
	}
	return string(out[0:indexes[0]]), nil
}

func main() {
	version := flag.Bool("v", false, "get tool version")
	autoPath := flag.Bool("a", false, "try get install path")
	flag.StringVar(&installPath, "i", installPath, "install path")
	flag.StringVar(&downloadPage, "d", downloadPage, "url for download page")
	flag.Parse()

	if *version {
		fmt.Println("go-updater v0.1")
		return
	}

	if *autoPath {
		var err error
		installPath, err = getInstallPath()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Will be installed to", installPath)
	}

	// Prepare to download
	fmt.Println("Get download URL")
	dl, err := getDownloadURL()
	if err != nil {
		fmt.Println(err)
		return
	}

	verInstalled, err := getGoVersionInstalled()
	if err != nil {
		fmt.Println(err)
		return
	}
	verDownload := getGoVersionActual(dl)

	if strings.Compare(verInstalled, verDownload) == 0 {
		fmt.Println("The current version is installed")
		return
	}

	// Download
	if err = downloadGo(dl); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Install to", installPath)
	if err = install(installPath, downloadPath); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Remove temp files")
	if err := os.Remove(downloadPath); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Update complete")
	out, err := exec.Command("bash", "-c", "go version").Output()
	fmt.Println(string(out))
	if err != nil {
		fmt.Println(err)
	}
}
