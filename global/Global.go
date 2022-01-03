package global

import (
	"net/http"
	"sync/atomic"
	"time"
)

var Debug bool
var seqNum int64 = 0

// HTTP client
var transport = &http.Transport{
	MaxIdleConns:    10,
	IdleConnTimeout: 30 * time.Second,
}

var Client = &http.Client{Transport: transport}

func NextSeq() int {
	return int(atomic.AddInt64(&seqNum, 1))
}
