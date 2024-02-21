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
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type config struct {
	Port    string            `toml:"port"`
	TLSPort string            `toml:"tls_port"`
	Cert    string            `toml:"cert"`
	Key     string            `toml:"key"`
	Domains map[string]string `toml:"domains"`
}

func (c config) isZero() bool {
	return c.Port == "" && c.TLSPort == "" && c.Cert == "" && c.Key == "" && len(c.Domains) == 0
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
				newcfg, err := loadConfig(path)
				if err != nil {
					log.Printf("error: failed to load config: %v", err)
					break
				}
				*cfgptr = newcfg
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			log.Println(err)
		}
	}
}

func loadConfig(path string) (cfg config, err error) {
	_, err = toml.DecodeFile(path, &cfg)
	return
}

func getConfig(path string) (cfg config) {
	flag.StringVar(&path, "cfg", path, "Path to the configuration file")
	flag.StringVar(&cfg.Port, "port", ":80", "The port that onering will listen to")
	flag.StringVar(&cfg.TLSPort, "tlsport", ":443", "The TLS port that onering will listen to")
	flag.StringVar(&cfg.Cert, "cert", "", "Path to the TLS certificate")
	flag.StringVar(&cfg.Key, "key", "", "Path to the TLS key")
	flag.Parse()

	cfg, err := loadConfig(path)
	if err != nil {
		log.Println(err)
	}

	if cfg.isZero() {
		log.Fatal("error: no configuration provided")
	}
	return cfg
}

func retry(d time.Duration, fn func()) {
	for {
		fn()
		time.Sleep(d)
	}
}

func main() {
	ucfg, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	cfgpath := filepath.Join(ucfg, "onering.toml")
	cfg = getConfig(cfgpath)
	go watch(cfgpath, &cfg)

	http.HandleFunc("/", redirect)

	go retry(time.Second*5, func() {
		log.Println(http.ListenAndServe(cfg.Port, nil))
	})
	retry(time.Second*5, func() {
		log.Println(http.ListenAndServeTLS(cfg.TLSPort, cfg.Cert, cfg.Key, nil))
	})
}
