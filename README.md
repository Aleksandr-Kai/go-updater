# Auto-update golang

**./go-updater -h**
```
Usage of ./go-updater:
  -d string
    	url for download page (default "https://go.dev/dl")
  -i string
    	install path (default "/usr/local")
  -v	get tool version
```

```
$ go version
go version go1.21.5 linux/amd64
$ ./go-updater
go: downloading github.com/PuerkitoBio/goquery v1.8.0
go: downloading github.com/dustin/go-humanize v1.0.0
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading golang.org/x/net v0.0.0-20210916014120-12bc252f5db8
Get download URL
Download file:  https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
Downloading... 69 MB complete      
Install to /usr/local
[sudo] password for <user_name>:       
Remove temp files
Update complete
go version go1.22.5 linux/amd64

$ go version
go version go1.22.5 linux/amd64
```
