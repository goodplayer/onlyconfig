package controller

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goodplayer/onlyconfig/webmgr/static"
)

func AddUserControllers(r *chi.Mux) {
	u := &UserController{}

	r.Post("/auth/user/login", u.UserLogin)

	//FIXME homepage
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		f, err := static.StaticFiles.ReadFile("files/index.html")
		if err != nil {
			log.Println("read file failed:", err)
			http.NotFound(writer, request)
			return
		}
		writer.Header().Add("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write(f); err != nil {
			log.Println("write http failed:", err)
		}
	})
}

type UserController struct {
}

func (u *UserController) UserLogin(w http.ResponseWriter, r *http.Request) {

}
