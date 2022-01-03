package http

import (
	"goproxy/config"
	"log"
	"net"
	"strings"
	"sync"
)

var mitmServerPool sync.Map

func ConnectRequest(clientConn net.Conn, data []byte) {
	log.Printf("ConnectRequest() %s\n", string(data))
	url := strings.Split(string(data), " ")[1]
	hostPort := strings.Split(url, ":")

	key := hostPort[0]
	mitmServer, ok := mitmServerPool.Load(key)
	if !ok {
		mitmServer = &MitmServer{
			protocol:       config.Https,
			host:           key,
			isForwardProxy: true,
			isSecure:       true,
			scheme:         "https",
		}
		mitmServer.(MitmServerInf).Add(1)
		loaded := false
		mitmServer, loaded = mitmServerPool.LoadOrStore(key, mitmServer)
		if loaded {
			mitmServer.(MitmServerInf).Wait()
		} else {
			log.Println("ConnectRequest() start https server")
			mitmServer.(MitmServerInf).Listen()
			mitmServer.(MitmServerInf).Done()
		}
	} else {
		log.Println("ConnectRequest() reuse https server")
		mitmServer.(MitmServerInf).Wait()
	}

	// Create tunnel from client to Http2HttpsServer
	createPipe(clientConn, mitmServer.(MitmServerInf).Address())
	sendConnectResponseToClient(clientConn)
}

func sendConnectResponseToClient(clientConn net.Conn) {
	log.Println("ConnectRequest respond() HTTP/1.1 200 Connection Established")
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n" +
		"Proxy-agent: GoProxy\r\n" +
		"\r\n"))
}
