package main

import (
	"allproxy/ca"
	"allproxy/global"
	"allproxy/http"
	"allproxy/paths"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func usage() {
	fmt.Println("\nUsage: allproxy [--listen [host:]port] [--debug]")
	fmt.Println("\nOptions:")
	fmt.Println("\t--listen - listen for incoming http connections.  Default is 8888.")
	fmt.Println("\nExample: allproxy --listen 8888")
}

type Listener struct {
	protocol string
	host     string
	port     string
}

const (
	httpX      = "httpx:"
	grpc       = "grpc:"
	secureGrpc = "securegrpc:"
)

func parseArgs() []Listener {
	listeners := make([]Listener, 0)
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--help":
			usage()
			os.Exit(1)
		case "--listen":
			if i+1 >= len(os.Args) {
				usage()
				fmt.Println("\nMissing port number for " + os.Args[i])
			}

			var protocol string = httpX
			switch os.Args[i] {
			case "--listen":
				protocol = httpX
			case "--listenGrpc":
				protocol = grpc
			case "--listenSecureGrpc":
				protocol = secureGrpc
			}
			var host string
			var port = os.Args[i]
			i++
			tokens := strings.Split(port, ":")
			if len(tokens) > 1 {
				host = tokens[0]
				port = tokens[1]
			}

			_, err := strconv.Atoi(port)
			if err != nil {
				usage()
				fmt.Println("\nInvalid port: " + os.Args[i])
				os.Exit(1)
			}
			listeners = append(listeners, Listener{protocol, host, port})
		case "--debug":
			global.Debug = true
		default:
			usage()
			fmt.Println("\nInvalid option: " + os.Args[i])
			os.Exit(1)
		}
	}
	return listeners
}

func main() {
	fmt.Println(os.Args)

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 	}
	// }()

	listeners := parseArgs()
	if len(listeners) == 0 {
		listeners = append(listeners, Listener{protocol: httpX, port: "8888"})
	}

	paths.MakeCaPemSymLink()

	ca.InitCa()

	for _, entry := range listeners {
		protocol := entry.protocol
		host := entry.host
		port := entry.port

		switch protocol {
		case httpX:
			fmt.Printf("Listening on %s %s %s\n", protocol, host, port)
			http.Listen(host + ":" + port)
			// case grpc:
			// 	GrpcProxy.forwardProxy(port, false)
			// 	console.log(`Listening on gRPC ${host || ''} ${port}`)
			// 	Global.portConfig.grpcPort = port
			// case secureGrpc:
			// 	GrpcProxy.forwardProxy(port, true)
			// 	console.log(`Listening on secure gRPC ${host || ''} ${port}`)
			// 	Global.portConfig.grpcSecurePort = port
		}
	}

}
