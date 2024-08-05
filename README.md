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
