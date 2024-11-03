
del /f/q onlyconfig.exe
del /f/q webmgr.exe
del /f/q onlyagent.exe

set GOOS=windows
set GOARCH=amd64

go build -o onlyconfig.exe ..
go build -o webmgr.exe ../cmd/webmgr
go build -o onlyagent.exe ../cmd/onlyagent
