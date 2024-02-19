# morpheus
A reverse proxy with a focus on simplicity.

### Example configuration file
```toml
# morpheus.toml

addr = ":8443"
cert = "/path/to/certificate.pem"
key = "/path/to/privatekey.pem"

[domains]
  "sub1.example.com" = "localhost:8080"
  "sub2.example.com" = "localhost:8081"
  "sub2.example.com" = "localhost:8082"
```
