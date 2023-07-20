# HTTPTLSClient

此项目基于github.com/refraction-networking/utls

基于 utls 功能开发实现自定义TLSExtensions

整合 net/http net/http2 实现RoundTripper接口, 封装新的HTTPClient, 使得utls兼容h1,h2

client用法与net/http一致, 详见example.go

```go
func New(proxy proxy.Proxy, helloSpec *utls.ClientHelloSpec) *http.Client{}
需要自定义ClientHelloSpec时, 可见utils.GetHelloSpec仿照一个传入 //nil则默认useragent类型
需要代理时,需自行实现Proxy接口逻辑,返回*url.URL //nil则不使用代理
或使用proxy.FuncToProxy把自定义函数转换成实现Proxy接口的函数
```





