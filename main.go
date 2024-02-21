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

type clicfg struct {
	path string
	config
}

func (c config) isZero() bool {
	return c.Port == "" && c.TLSPort == "" && c.Cert == "" && c.Key == "" && len(c.Domains) == 0
}

func (c *config) override(n config) {
	if n.Port != "" {
		c.Port = n.Port
	}
	if n.TLSPort != "" {
		c.TLSPort = n.TLSPort
	}
	if n.Cert != "" {
		c.Cert = n.Cert
	}
	if n.Key != "" {
		c.Cert = n.Cert
	}
	if n.Domains != nil {
		c.Domains = n.Domains
	}
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
				if _, err := toml.DecodeFile(path, cfgptr); err != nil {
					log.Println("watch", "toml.DecodeFile", err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			log.Println("watch", err)
		}
	}
}

func parseFlags() (c clicfg) {
	flag.StringVar(&c.path, "cfg", "", "Path to the configuration file")
	flag.StringVar(&c.Port, "port", "", "The port that onering will listen to")
	flag.StringVar(&c.TLSPort, "tlsport", "", "The TLS port that onering will listen to")
	flag.StringVar(&c.Cert, "cert", "", "Path to the TLS certificate")
	flag.StringVar(&c.Key, "key", "", "Path to the TLS key")
	flag.Parse()

	return
}

func getConfig(path string) (c config) {
	cli := parseFlags()

	if cli.path != "" {
		path = cli.path
	}

	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		log.Println("getConfig", "toml.DecodeFile", err)
	}

	c.override(cli.config)
	if c.isZero() {
		log.Fatal("error: no configuration provided")
	}
	return
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
		log.Fatal("main", "os.UserConfigDir", err)
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
