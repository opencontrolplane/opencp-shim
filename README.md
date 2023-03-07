# opencp-shim
OpenCP shim is a simple HTTP server that implements the Kubernetes API server interface. It is a shim that allows you to use the Kubernetes API server to implement your own API server.

## How to run it in development mode
if this is for development, you can run it with the following steps:

* Clone the repository
* Run `go build`
* Export `SSL`, this is to use the local SSL certificates
* Run `./opencp-shim`

### Using Docker
Build using the following command:
```bash
docker build -t opencp-shim .
```
Run it using the following command:
```bash
docker run -d -p 4000:4000 -e SSL=true opencp-shim
```
## How to run it in production mode
if this is for production, you dont need to build it, you can run it with the following steps:
```bash
docker run -d -p 4000:4000 opencp-shim
```
but you need a valid SSL certificate in front of it. You can use [certbot](https://certbot.eff.org/) to get a valid SSL certificate.
or a valid SSL certificate from a CA. Also you can use [nginx](https://www.nginx.com/) to proxy the traffic to the opencp-shim or caddy or any other reverse proxy.
