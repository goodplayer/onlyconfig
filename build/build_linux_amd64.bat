
del /f/q onlyconfig.exe
del /f/q webmgr.exe

set GOOS=linux
set GOARCH=amd64

go build -o onlyconfig ..
go build -o webmgr ../cmd/webmgr
