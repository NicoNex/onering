package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type config struct {
	path    string
	Port    string            `toml:"port"`
	TLSPort string            `toml:"tls_port"`
	Cert    string            `toml:"cert"`
	Key     string            `toml:"key"`
	Domains map[string]string `toml:"domains"`
}

func parseFlags(cfg *config) error {
	raw := flag.Args()
	clean := make([]string, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case "-d", "--domain":
			if i+2 >= len(raw) {
				return fmt.Errorf("%s flag requires an origin and a target", raw[i])
			}
			origin, target := raw[i+1], raw[i+2]
			if cfg.Domains == nil {
				cfg.Domains = make(map[string]string)
			}
			cfg.Domains[origin] = target
			i += 2
		default:
			clean = append(clean, raw[i])
		}
	}

	fs := flag.NewFlagSet("onering", flag.ExitOnError)
	fs.Usage = usage

	var help bool
	fs.BoolVar(&help, "h", false, "Show help")
	fs.BoolVar(&help, "help", false, "Show help")

	fs.StringVar(&cfg.Port, "port", cfg.Port, "The port that onering will listen to")
	fs.StringVar(&cfg.TLSPort, "tlsport", cfg.TLSPort, "The TLS port that onering will listen to")
	fs.StringVar(&cfg.Cert, "cert", cfg.Cert, "Path to the TLS certificate")
	fs.StringVar(&cfg.Key, "key", cfg.Key, "Path to the TLS key")

	if err := fs.Parse(clean); err != nil {
		return err
	}
	if help {
		usage()
		os.Exit(0)
	}
	return nil
}

func defaultConfigPath() string {
	cfgpath, _ := os.UserConfigDir()
	return filepath.Join(cfgpath, "onering.toml")
}

func parseConfigFlag(defaultPath string) string {
	fs := flag.NewFlagSet("cfg", flag.ContinueOnError)
	fs.Usage = usage

	var help bool
	fs.BoolVar(&help, "h", false, "Show help")
	fs.BoolVar(&help, "help", false, "Show help")

	fs.StringVar(&defaultPath, "cfg", defaultPath, "Path to the configuration file")
	fs.Parse(os.Args[1:])

	if help {
		usage()
		os.Exit(0)
	}
	return defaultPath
}

func configurations() (cfg config, err error) {
	flag.Usage = usage
	cfg.path = parseConfigFlag(defaultConfigPath())
	cfg.Port, cfg.TLSPort = ":8080", ":443"

	if _, e := toml.DecodeFile(cfg.path, &cfg); e != nil {
		err = errors.Join(err, e)
	}
	return cfg, errors.Join(err, parseFlags(&cfg))
}

func usage() {
	fmt.Printf(`onering — A reverse proxy with a focus on simplicity.

Usage:
  %s [options]

Options:
  -cfg string               Path to the configuration file. (default "%s")
  -port string              The port on which onering will listen. (default ":8080")
  -tlsport string           The TLS port on which onering will listen. (default ":443")
  -cert string              Path to the TLS certificate. Required for TLS.
  -key string               Path to the TLS key. Required for TLS.
  -d, --domain origin target
                            Add a domain mapping: requests for ORIGIN
                            will be proxied to TARGET. Can be repeated.
  -h, --help                Show this help message and exit.

Examples:
  %s
    Run onering with the default config file and ports.

  %s -cfg /etc/onering/config.toml
      Load configuration from /etc/onering/config.toml.

  %s -port :9090 -tlsport :9443 -cert cert.pem -key key.pem
      Override ports and TLS certificate/key on the command line.

  %s \
    -d https://foo.example.com https://localhost:8081 \
    -d https://bar.example.com https://localhost:8082
        Proxy foo→localhost:8081 and bar→localhost:8082.

onering © 2024  Nicolò Santamaria
This program comes with ABSOLUTELY NO WARRANTY.
`,
		os.Args[0],
		defaultConfigPath(),
		os.Args[0],
		os.Args[0],
		os.Args[0],
		os.Args[0],
	)
}
