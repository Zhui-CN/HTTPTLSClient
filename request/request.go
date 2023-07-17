package request

import (
	"HTTPTLSClient/proxy"
	"HTTPTLSClient/utils"
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	dialTimeout           = time.Second * 15
	clientTimeout         = time.Second * 60
	proxyAuthHead         = "Proxy-Authorization"
	proxyCtxKey           = "proxy"
	tlsConnCtxKey         = "tlsConn"
	clientHelloSpecCtxKey = "clientHelloSpec"
)

// 实现 http RoundTripper接口
type uTransport struct {
	tr1             *http.Transport       // h1 Transport
	tr2             *http2.Transport      // h2 Transport
	alpnMap         map[string]string     // 保存alpn
	proxy           proxy.Proxy           // 自定义proxy
	clientHelloSpec *utls.ClientHelloSpec // 可自定义clientHelloSpec 如果为nil, 则用useragent或chrome对应的拓展
}

// CloseIdleConnections 实现接口关闭连接
func (u *uTransport) CloseIdleConnections() {
	u.tr1.CloseIdleConnections()
	u.tr2.CloseIdleConnections()
}

// RoundTrip 实现接口预处理请求逻辑
func (u *uTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 如果不是https的请求, 直接使用默认逻辑
	if req.URL.Scheme != "https" {
		fmt.Println("当前请求使用协议: ", "h1")
		return u.tr1.RoundTrip(req)
	}
	// 处理https
	ctx := req.Context()
	if u.proxy != nil {
		if proxyURL := u.proxy.GetURL(req); proxyURL != nil {
			ctx = context.WithValue(ctx, proxyCtxKey, proxyURL)
		}
	}
	if u.clientHelloSpec != nil {
		ctx = context.WithValue(ctx, clientHelloSpecCtxKey, u.clientHelloSpec)
	} else {
		ctx = context.WithValue(ctx, clientHelloSpecCtxKey, utils.GetClientHelloSpec(req.UserAgent()))
	}
	addr := utils.GetURLAddr(req.URL)
	alpn, ok := u.alpnMap[addr]
	if !ok {
		// 未知的alpn,则先预处理,并用Context保存conn供后续使用,不重复创建连接
		tlsConn, err := getTLSConn(ctx, addr)
		if err != nil {
			return nil, err
		}
		// 用Context保存conn供后续连接阶段直接返回
		ctx = context.WithValue(ctx, tlsConnCtxKey, tlsConn)
		alpn = tlsConn.ConnectionState().NegotiatedProtocol
		u.alpnMap[addr] = alpn
	}
	req = req.WithContext(ctx)
	fmt.Println("当前请求使用协议: ", alpn)
	if alpn == "h2" {
		return u.tr2.RoundTrip(req)
	}
	return u.tr1.RoundTrip(req)
}

// CONNECT Proxy
func httpsConnectToProxy(proxyURL *url.URL, addr string) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", utils.GetURLAddr(proxyURL), dialTimeout)
	connectReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	if auth := utils.ProxyAuth(proxyURL); auth != "" {
		connectReq.Header.Set(proxyAuthHead, auth)
	}
	var resp *http.Response
	if err = connectReq.Write(conn); err != nil {
		return
	}
	if resp, err = http.ReadResponse(bufio.NewReader(conn), connectReq); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		err = errors.New("proxy refused connection" + string(body))
	}
	return
}

// 建立TLS连接
func getTLSConn(ctx context.Context, addr string) (tlsConn *utls.UConn, err error) {
	// 如果已预处理获取过conn,则直接返回
	if c := ctx.Value(tlsConnCtxKey); c != nil {
		return c.(*utls.UConn), nil
	}
	var conn net.Conn
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	// 如果设置了代理,建立隧道
	if p := ctx.Value(proxyCtxKey); p != nil {
		conn, err = httpsConnectToProxy(p.(*url.URL), addr)
	} else {
		conn, err = net.DialTimeout("tcp", addr, dialTimeout)
	}
	host, _, _ := net.SplitHostPort(addr)
	tlsConn = utls.UClient(conn, &utls.Config{ServerName: host, InsecureSkipVerify: true}, utls.HelloCustom)
	spec := ctx.Value(clientHelloSpecCtxKey).(*utls.ClientHelloSpec)
	err = tlsConn.ApplyPreset(spec)
	err = tlsConn.Handshake()
	return
}

// HTTPS H1 DialTLS
func tr1DialTLS(ctx context.Context, _, addr string) (net.Conn, error) {
	return getTLSConn(ctx, addr)
}

// HTTPS H2 DialTLS
func tr2DialTLS(ctx context.Context, _, addr string, _ *tls.Config) (net.Conn, error) {
	return getTLSConn(ctx, addr)
}

// New HTTP Client
func New(proxy proxy.Proxy, helloSpec *utls.ClientHelloSpec) *http.Client {
	tr1 := http.DefaultTransport.(*http.Transport).Clone()
	tr1.DialTLSContext = tr1DialTLS
	if proxy != nil {
		tr1.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxy.GetURL(req), nil
		}
	}
	tr2 := &http2.Transport{
		DialTLSContext:            tr2DialTLS,
		MaxHeaderListSize:         262144,
		MaxDecoderHeaderTableSize: 65536,
	}
	return &http.Client{
		Transport: &uTransport{
			tr1:             tr1,
			tr2:             tr2,
			alpnMap:         make(map[string]string),
			proxy:           proxy,
			clientHelloSpec: helloSpec,
		},
		Timeout: clientTimeout,
	}
}
