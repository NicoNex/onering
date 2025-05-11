package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type rproxy map[string]*url.URL

func newRProxy(domains map[string]string) (rp rproxy, err error) {
	rp = make(rproxy)

	for origin, target := range domains {
		u, e := url.Parse(target)
		if e != nil {
			errors.Join(err, e)
			continue
		}
		rp[origin] = u
	}
	return
}

func (rp *rproxy) redirect(w http.ResponseWriter, r *http.Request) {
	origin, ok := (*rp)[r.Host]
	if !ok {
		http.NotFound(w, r)
		return
	}
	httputil.NewSingleHostReverseProxy(origin).ServeHTTP(w, r)
}

func watch(path string, rp *rproxy) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("watch", "fsnotify.NewWatcher", err)
	}
	defer watcher.Close()
	if err := watcher.Add(filepath.Dir(path)); err != nil {
		log.Println("watch", "watcher.Add", err)
		return
	}

loop:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue loop
			}
			if event.Has(fsnotify.Write) && event.Name == path {
				var cfg config
				if _, err := toml.DecodeFile(path, &cfg); err != nil {
					log.Println("watch", "toml.DecodeFile", err)
				}

				newrp, err := newRProxy(cfg.Domains)
				if err != nil {
					log.Println("watch", "newRProxy", err)
				}
				*rp = newrp
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			log.Println("watch", err)
		}
	}
}

func retry(d time.Duration, fn func()) {
	for {
		fn()
		time.Sleep(d)
	}
}

func main() {
	cfg, err := configurations()
	if err != nil {
		log.Println("main", "configurations", err)
	}

	rp, err := newRProxy(cfg.Domains)
	if err != nil {
		log.Println("main", "newRproxy", err)
	}

	go watch(cfg.path, &rp)
	http.HandleFunc("/", rp.redirect)

	go retry(time.Second*5, func() {
		log.Println(
			"main",
			"http.ListenAndServe",
			http.ListenAndServe(cfg.Port, nil),
		)
	})
	retry(time.Second*5, func() {
		log.Println(
			"main",
			"http.ListenAndServeTLS",
			http.ListenAndServeTLS(cfg.TLSPort, cfg.Cert, cfg.Key, nil),
		)
	})
}
