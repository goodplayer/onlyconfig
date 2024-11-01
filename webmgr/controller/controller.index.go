package controller

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/goodplayer/onlyconfig/webmgr/static"
)

func (cc *ControllerContainer) addIndex(r *chi.Mux) {

	// homepage
	f := func(writer http.ResponseWriter, request *http.Request) {
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
	}
	r.Get("/", f)
	// webmgr_project routers
	r.Get("/login", f)
	r.Get("/logout", f)
	r.Get("/register", f)
	r.Get("/change_password", f)
	r.Get("/env_and_dc", f)
	r.Get("/org_mgr", f)

	// static files
	{
		scriptsFs, err := fs.Sub(static.StaticFiles, "files")
		if err != nil {
			log.Println("sub scripts filesystem failed:", err)
			panic(err)
		}
		r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
			var fp = r.RequestURI
			if strings.HasPrefix(fp, "/") {
				fp = fp[1:]
			}
			http.FileServer(http.FS(scriptsFs)).ServeHTTP(w, r)
		})
	}

	//FIXME default router: set index.html preact project container as default router
	{
	}

	//FIXME add additional files including .svg and so on
	{
	}
}
