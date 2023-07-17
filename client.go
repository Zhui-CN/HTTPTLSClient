package HTTPTLSClient

import (
	"HTTPTLSClient/proxy"
	"HTTPTLSClient/request"
	utls "github.com/refraction-networking/utls"
	"net/http"
)

// New HTTP Client
func New(proxy proxy.Proxy, helloSpec *utls.ClientHelloSpec) *http.Client {
	return request.New(proxy, helloSpec)
}
