package main

import (
	"fmt"
	"github.com/Zhui-CN/HTTPTLSClient"
	"github.com/Zhui-CN/HTTPTLSClient/proxy"
	"github.com/Zhui-CN/HTTPTLSClient/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	resp   *http.Response
	req, _ = http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
)

func DumpResponseNoBody(response *http.Response) {
	defer response.Body.Close()
	resp, err := httputil.DumpResponse(response, true)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to dump response: %v", err))
	}
	fmt.Println(string(resp))
}

func do(resp *http.Response, err error) *http.Response {
	if err != nil {
		panic(err)
	}
	return resp
}

func getProxy(req *http.Request) *url.URL {
	return nil
}

func mainWithProxy() *http.Response {
	client := HTTPTLSClient.New(proxy.FuncToProxy(getProxy), nil)
	req.Header.Set(proxy.TypeKey, "default")
	req.Header.Set(proxy.CountryKey, "us")
	defer client.CloseIdleConnections()
	return do(client.Do(req))
}

func mainWithSpec() *http.Response {
	spec := utils.GetHelloSpec()
	utils.ShuffleExtensions(spec)
	client := HTTPTLSClient.New(nil, spec)
	defer client.CloseIdleConnections()
	return do(client.Do(req))
}

func main() {
	//resp = mainWithProxy()
	resp = mainWithSpec()
	DumpResponseNoBody(resp)
}
