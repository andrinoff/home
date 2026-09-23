// Command home is the home manager server: a single binary serving the
// embedded React frontend, the JSON API and the SQLite database.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/andrinoff/home/internal/api"
	"github.com/andrinoff/home/internal/store"
	"github.com/andrinoff/home/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	data := flag.String("data", "./data", "directory for the SQLite database")
	flag.Parse()

	if err := os.MkdirAll(*data, 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	st, err := store.Open(filepath.Join(*data, "home.db"))
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	static, err := web.FS()
	if err != nil {
		log.Fatalf("load embedded frontend: %v", err)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.NewServer(st, static),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("home manager listening on %s (data dir: %s)", *addr, *data)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
