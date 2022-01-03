<h1 align="center" style="border-bottom: none;">Go HTTP(s) Forward/Reverse Proxy</h1>

This is intended to provide the proxy for the goproxy frontend.  It is currently a work in progress, and is not very stable, and not fully functional.

## Setup
1. Set path to data directory:
```sh
goproxy$ export GOPROXY_DATA_DIR=$HOME/goproxy-data       # path where certificates and keys are stored
```
2. Start *goproxy*:
```sh
goproxy$ go run goproxy
```
3. Import *$GOPROXY_DATA_DIR/ca.pem* into your browser trust store.
4. Configure your browser to proxy https and http to host *localhost* and port *8888*.

## License

This code is licensed under the [MIT License](https://opensource.org/licenses/MIT).
