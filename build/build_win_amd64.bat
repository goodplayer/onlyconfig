
del /f/q onlyconfig.exe
del /f/q webmgr.exe

set GOOS=windows
set GOARCH=amd64

go build -o onlyconfig.exe ..
go build -o webmgr.exe ../cmd/webmgr
