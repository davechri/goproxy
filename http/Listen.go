package http

import (
	"bufio"
	"goproxy/api"
	"goproxy/config"
	"goproxy/paths"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	socketio "github.com/googollee/go-socket.io"
)

var socketioServer *socketio.Server
var mitmHttpsServer = &MitmServer{
	protocol:       config.Https,
	host:           "goproxy",
	isForwardProxy: false,
	isSecure:       true,
	scheme:         "https",
} // secure reverse proxy
var mitmHttpServer = &MitmServer{
	protocol:       config.Http,
	host:           "goproxy",
	isForwardProxy: false,
	isSecure:       false,
	scheme:         "http",
} // secure reverse proxy

func Listen(address string) {
	log.Printf("Listen(%s)\n", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Panicln(err)
	}

	temp := api.Start()
	socketioServer = temp

	// This is causing issue with inbound connection handling
	// http.HandleFunc("/", DashboardAppServer)
	// go http.Serve(listener, nil)

	// Setup https and http reverse proxy servers
	mitmHttpsServer.Listen()
	mitmHttpServer.Listen()

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicln(err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	log.Printf("Listen handleRequest(%v)\n", conn)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Listen handleRequest() error %v\n%s\n", err, string(buf))
		return
	}

	if n >= len("CONNECT") && strings.HasPrefix(string(buf), "CONNECT") {
		log.Printf("Listen handleRequest() CONNECT\n")
		ConnectRequest(conn, buf)
	} else {
		log.Printf("Listen handleRequest() %v\n", string(buf))
		if isClientHello(buf) {
			log.Printf("Listen handleRequest() client hello: %s\n", string(buf))
			conn, err := net.Dial("tcp", mitmHttpsServer.Address())
			if err != nil {
				log.Panicln(err)
			}
			_, err = conn.Write(buf)
			if err != nil {
				log.Panicln(err)
			}
			createPipe(conn, mitmHttpsServer.Address())
		} else { // Assume this is just HTTP in the clear
			log.Println("Listen handleRequest() http:\n", string(buf))

			rdr := bufio.NewReader(strings.NewReader(string(buf)))
			request, err := http.ReadRequest(rdr)
			if err != nil {
				log.Panicln(err)
			}

			header := http.Header{}
			headersWritten := false
			responseWriter := connResponseWriter{conn, header, &headersWritten}

			dir := filepath.Join(paths.ClientDir(), "build")
			file := filepath.Join(dir, request.URL.Path)

			if request.URL.Path == "/socket.io/" {
				log.Println("Listen handleRequest() socket.io", request.URL.Host, request.URL.Path)
				socketioServer.ServeHTTP(responseWriter, request)
			} else if _, err := os.Stat(file); err == nil {
				fs := http.FileServer(http.Dir(dir))
				fs.ServeHTTP(responseWriter, request)
			} else {
				conn, err := net.Dial("tcp", mitmHttpServer.Address())
				if err != nil {
					log.Panicln(err)
				}
				_, err = conn.Write(buf)
				if err != nil {
					log.Panicln(err)
				}
				createPipe(conn, mitmHttpServer.Address())
			}
		}
	}
}

func isClientHello(buf []byte) bool {
	return len(buf) >= 3 &&
		buf[0] == 0x16 &&
		buf[1] == 0x03 &&
		buf[2] == 0x01
}
