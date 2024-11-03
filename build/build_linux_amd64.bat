
del /f/q onlyconfig
del /f/q webmgr
del /f/q onlyagent

set GOOS=linux
set GOARCH=amd64

go build -o onlyconfig ..
go build -o webmgr ../cmd/webmgr
go build -o onlyagent ../cmd/onlyagent
