package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/meidoworks/nekoq-component/configure/configserver"

	"github.com/goodplayer/onlyconfig/datapump"
)

var httpAddr string
var postgresAddr string

func init() {
	//FIXME default ports: read-http=8800, read-https=8801, write-http=8802, write-https=8803
	flag.StringVar(&httpAddr, "http", ":8800", "http listen address")
	flag.StringVar(&postgresAddr, "postgres", "postgres://admin:admin@127.0.0.1:5432/onlyconfig", "postgres connection string")
	flag.Parse()
}

func main() {
	dp := datapump.NewPostgresDataPump(postgresAddr)

	opt := configserver.ConfigureOptions{
		Addr: httpAddr,
		TLSConfig: struct {
			Addr string
			Cert string
			Key  string
		}{},
		MaxWaitTimeForUpdate: 60,
		DataPump:             dp,
	}
	server := configserver.NewConfigureServer(opt)
	if err := server.Startup(); err != nil {
		panic(err)
	}
	defer func(server *configserver.ConfigureServer) {
		err := server.Shutdown()
		if err != nil {
			log.Println("error while shutting down server ", err)
		}
	}(server)

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	log.Println("Shutting down...")
}
