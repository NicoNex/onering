package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type config struct {
	Addr    string            `toml:"addr"`
	Cert    string            `toml:"cert"`
	Key     string            `toml:"key"`
	Domains map[string]string `toml:"domains"`
}

func (c config) isZero() bool {
	return c.Addr == "" && c.Cert == "" && c.Key == "" && len(c.Domains) == 0
}

var cfg config

func redirect(w http.ResponseWriter, r *http.Request) {
	target, ok := cfg.Domains[r.Host]
	if !ok {
		fmt.Fprintf(w, "error: not found")
		return
	}

	origin, err := url.Parse(target)
	if err != nil {
		fmt.Fprintf(w, "error: failed to parse url")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(origin)
	proxy.ServeHTTP(w, r)
}

func watch(path string, cfgptr *config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	watcher.Add(path)

loop:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue loop
			}
			if event.Has(fsnotify.Write) {
				cfg, err := loadConfig(path)
				if err != nil {
					log.Printf("error: failed to load config: %v", err)
					continue loop
				}
				*cfgptr = cfg
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				continue loop
			}
			log.Println(err)
		}
	}
}

func loadConfig(path string) (cfg config, err error) {
	_, err = toml.DecodeFile(path, &cfg)
	return
}

func getConfig(path string) config {
	cfg, err := loadConfig(path)
	if err != nil {
		log.Println(err)
	}

	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "The address that morpheus will listen to")
	flag.StringVar(&cfg.Cert, "cert", cfg.Cert, "Path to the TLS certificate")
	flag.StringVar(&cfg.Key, "key", cfg.Key, "Path to the TLS key")
	flag.Parse()

	if cfg.isZero() {
		log.Fatal("error: no configuration provided")
	}
	return cfg
}

func main() {
	ucfg, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	cfgpath := filepath.Join(ucfg, "morpheus.toml")
	cfg = getConfig(cfgpath)
	go watch(cfgpath, &cfg)

	http.HandleFunc("/", redirect)
	log.Fatal(http.ListenAndServe(cfg.Addr, nil))
}
