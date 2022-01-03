package http

import (
	"goproxy/paths"
	"net/http"
	"os"
	"path/filepath"
)

func DashboardAppServer(w http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/goproxy" {
		http.Redirect(w, request, "/", http.StatusMovedPermanently)
	}

	dir := filepath.Join(paths.ClientDir(), "build")
	file := filepath.Join(dir, request.URL.Path)

	if _, err := os.Stat(file); err == nil {
		fs := http.FileServer(http.Dir(dir))
		fs.ServeHTTP(w, request)
	}
}
