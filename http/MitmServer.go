package http

import (
	"allproxy/api"
	"allproxy/ca"
	"allproxy/config"
	"allproxy/dns"
	"allproxy/global"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"sync/atomic"
)

const allproxySeqHeader = "allproxy-seq" // add "allproxy-seq" to request header

type MitmServerInf interface {
	Listen()
	Address() string
	Add(int)
	Wait()
	Done()
}

type MitmServer struct {
	protocol            config.ConfigProtocol
	host                string
	port                int
	isForwardProxy      bool
	isSecure            bool
	scheme              string
	waitGroup           sync.WaitGroup
	pipelineSeqNum      int32                // Use atomic.AddUint32()
	seqToHttpMessageMap map[int]*HttpMessage // lookup HttpMessage key=seq num
	reverseProxy        *httputil.ReverseProxy
}

func (s *MitmServer) Listen() {
	log.Printf("MitmServer Listen() Listen %v\n", s)

	s.seqToHttpMessageMap = make(map[int]*HttpMessage)

	addr := "localhost:0"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicln(err)
	}
	s.port = listener.Addr().(*net.TCPAddr).Port
	log.Printf("MitmServer Listen() Listen on port %d %v", s.port, s)

	// Set up reverse proxy
	proxy := &httputil.ReverseProxy{}
	proxy.Director = func(request *http.Request) {
		seqNum, _ := strconv.Atoi(request.Header.Get(allproxySeqHeader))
		httpMessage := s.seqToHttpMessageMap[seqNum]
		request.URL.Scheme = s.scheme
		host := s.host
		if !s.isForwardProxy {
			host = httpMessage.ProxyConfig.Hostname
		}
		request.URL.Host = host
		request.Header.Set("host", host)
	}
	proxy.ModifyResponse = func(res *http.Response) error {
		return s.responseHandler(res)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Panicln(err)
	}
	s.reverseProxy = proxy

	// Start serving HTTP requests
	mux := http.NewServeMux()
	mux.Handle("/", s)
	if s.isSecure {
		certFile, keyFile := ca.NewServerCertKey(s.host)
		go http.ServeTLS(listener, mux, certFile, keyFile)
	} else {
		go http.Serve(listener, mux)
	}
}

func (s *MitmServer) Add(delta int) {
	log.Printf("MitmServer Add(%d) %v\n", delta, s)
	s.waitGroup.Add(delta)
}

func (s *MitmServer) Wait() {
	log.Printf("MitmServer Wait() %v\n", s)
	s.waitGroup.Wait()
	log.Printf("MitmServer Wait() done %v\n", s)
}

func (s *MitmServer) Done() {
	log.Printf("MitmServer Done() %v\n", s)
	s.waitGroup.Done()
}

func (s *MitmServer) Address() string {
	log.Printf("MitmServer Address() %d %v\n", s.port, s)
	return "localhost:" + strconv.FormatInt(int64(s.port), 10)
}

// HTTP request handler
func (s *MitmServer) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	pipelineSeqNum := atomic.AddInt32(&s.pipelineSeqNum, 1)
	globalSeqNum := global.NextSeq()
	log.Printf("MitmServer captureRequest() seq=%d\n", globalSeqNum)

	// Find matching proxy configuration
	clientHostName := dns.ResolveIp(request.RemoteAddr)
	proxyConfig := api.FindProxyConfigMatchingURL(s.isSecure, clientHostName, request.URL, s.isForwardProxy)
	// Always proxy forward proxy requests
	if proxyConfig == nil && s.isForwardProxy {
		proxyConfig = &config.ProxyConfig{
			IsSecure:      s.isSecure,
			Path:          request.URL.Path,
			Protocol:      s.protocol,
			Hostname:      s.host,
			Port:          s.port,
			HostReachable: true,
			Comment:       "Created by AllProxy",
		}
	}

	if proxyConfig == nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<h1>No AllProxy config is defined for path: " + request.URL.Path + "</h1>"))
		return
	}

	// read all bytes from content body and create new stream using it.
	var reqBody []byte
	var err error
	if request.Body != nil {
		reqBody, err = io.ReadAll(request.Body)
		if err != nil {
			log.Panicln(err)
		}
		request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	messageProtocol := api.Https
	if !s.isSecure {
		messageProtocol = api.Http
	}

	httpMessage := NewHttpMessage(
		messageProtocol,
		proxyConfig,
		pipelineSeqNum,
		globalSeqNum,
		request.RemoteAddr,
		request.Method,
		request.URL.String(),
		request.Header,
		reqBody,
	)

	httpMessage.EmitMessageToBrowser(
		0,
		nil,
		api.NoResponse,
	)
	s.seqToHttpMessageMap[globalSeqNum] = httpMessage
	request.Header.Set(allproxySeqHeader, strconv.Itoa(int(globalSeqNum)))
	s.reverseProxy.ServeHTTP(w, request)
}

// HTTP Response handler
func (s *MitmServer) responseHandler(res *http.Response) error {
	seqNum, _ := strconv.Atoi(res.Request.Header.Get(allproxySeqHeader))
	httpMessage := s.seqToHttpMessageMap[seqNum]
	delete(s.seqToHttpMessageMap, seqNum)
	var resBody []byte
	if res.Body != nil {
		var err error
		resBody, err = io.ReadAll(res.Body)
		if err != nil {
			log.Panicln(err)
		}
		res.Body = io.NopCloser(bytes.NewBuffer(resBody))
	}
	httpMessage.EmitMessageToBrowser(
		res.StatusCode,
		res.Header,
		resBody,
	)
	return nil
}