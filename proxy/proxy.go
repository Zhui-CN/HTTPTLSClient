package proxy

import (
	"fmt"
	"net/http"
	"net/url"
)

var (
	TypeKey    = "proxy-type"
	CountryKey = "proxy-country"
)

// Proxy 实现获取对应proxy逻辑的接口
type Proxy interface {
	GetURL(*http.Request) *url.URL
}

// GetURL 实现Proxy接口
type proxyMap map[string]*url.URL

// GetURL 实现GetURL,获取自定义处理后的proxyURL
func (m proxyMap) GetURL(req *http.Request) (proxyURL *url.URL) {
	header := req.Header
	proxyType := header.Get(TypeKey)
	proxyCountry := header.Get(CountryKey)
	req.Header.Del(TypeKey)
	req.Header.Del(CountryKey)
	if u, ok := m[proxyType]; ok {
		if proxyCountry != "" {
			proxyCountry = "-country-" + proxyCountry
		}
		username := fmt.Sprintf(u.User.Username(), proxyCountry)
		pwd, _ := u.User.Password()
		proxyURL = &url.URL{
			User: url.UserPassword(username, pwd),
			Host: u.Host,
		}
	}
	return
}

// Proxies 自定义proxy
var Proxies = proxyMap{
	"default": &url.URL{
		User: url.UserPassword("xxx", "xxx"),
		Host: "xxxxxxx:8888",
	},
}
