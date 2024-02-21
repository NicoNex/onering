# OneRing: Simple and Secure Reverse Proxy

**OneRing** is a robust and user-friendly reverse proxy designed to simplify subdomain management through a TOML configuration file. It not only eases the management of subdomains but also takes the hassle out of dealing with HTTPS (TLS) certificates for each backend service it points to. With OneRing, you can work seamlessly as if you were working locally.

## Key Features:

- **Dynamic Subdomain Creation:** Easily manage subdomains through a straightforward TOML configuration file. The example TOML file allows you to map subdomains to specific backend services effortlessly.

- **High Performance:** OneRing is built to handle a large number of requests, scaling with your hardware to deliver optimal performance.

- **Security:** With OneRing, security is a top priority. The reverse proxy provides a secure environment for your applications, and it takes care of establishing encrypted connections on its own.

- **Simplified TLS Management:** Although OneRing requires you to provide TLS certificates and keys, it exempts you from dealing with them for each backend service. OneRing acts as the one ring to rule them all, handling HTTPS intricacies for you and allowing you to work as if you were working locally.

- **Simple Configuration:** Configuring OneRing is a breeze. Modify the TOML configuration file, and the changes take effect instantly, without the need to restart the proxy. Place the configuration file (`onering.toml`) in the user config directory for seamless integration with OneRing.

## Example TOML Configuration:

```toml
port = ":80"
tls_port = ":443"
cert = "/path/to/certificate.pem"
key = "/path/to/privatekey.pem"

[domains]
  "example.com" = "http://localhost:8080"
  "sub1.example.com" = "http://localhost:8081"
  "sub2.example.com" = "http://localhost:8082"
  "sub3.example.com" = "http://localhost:8083"
```

## Command-Line Interface (CLI) Flags:

- **-cfg:** Path to the configuration file (default: user config directory/onering.toml).
- **-port:** The port that OneRing will listen to.
- **-tlsport:** The TLS port that OneRing will listen to.
- **-cert:** Path to the TLS certificate.
- **-key:** Path to the TLS key.

## Example Usage:

If the configuration file (`onering.toml`) is in the right location, simply run:

```bash
./onering
```

For a real-world scenario, ensuring continuous operation and logging, you might want to run it using:

```bash
nohup ./onering &>> /path/to/logfile.log &
```

Alternatively if you prefer specifying configuration via CLI flags, use the following command:

```bash
./onering -port :8080 -tlsport :8443 -cert /path/to/cert.pem -key /path/to/key.pem
```
