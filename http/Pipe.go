package http

import (
	"io"
	"log"
	"net"
)

// Create tunnel from client to goproxy https server.  The goproxy https server decrypts and captures
// the HTTP messages, and forwards it to the origin server.
func createPipe(clientConn net.Conn, address string) {
	log.Printf("Pipe createPipe(%s)\n", address)
	serverConn, err := net.Dial("tcp", address)
	if err != nil {
		log.Panicln(err)
	}

	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			log.Println(err)
		}
		clientConn.Close()
		serverConn.Close()
	}()

	go func() {
		_, err := io.Copy(clientConn, serverConn)
		if err != nil {
			log.Println(err)
		}
		clientConn.Close()
		serverConn.Close()
	}()
}
