package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"xorm.io/xorm"

	"github.com/goodplayer/onlyconfig/webmgr/controller"
)

var debugFlag bool
var httpAddr string
var postgresAddr string

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "debug mode")
	flag.StringVar(&httpAddr, "http", ":8880", "http listen address")
	flag.StringVar(&postgresAddr, "postgres", "postgres://admin:admin@127.0.0.1:5432/onlyconfig", "postgres connection string")
	flag.Parse()
}

func main() {

	engine, err := xorm.NewEngine("pgx", postgresAddr)
	if err != nil {
		panic(err)
	}
	if debugFlag {
		engine.ShowSQL(true)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	controller.AddUserControllers(r)

	if err := http.ListenAndServe(httpAddr, r); err != nil {
		panic(err)
	}

}
