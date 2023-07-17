package utils

import (
	"encoding/base64"
	"errors"
	"github.com/mssola/useragent"
	utls "github.com/refraction-networking/utls"
	"math/rand"
	"net"
	"net/url"
)

func GetURLAddr(url *url.URL) string {
	host := url.Hostname()
	port := url.Port()
	if port == "" {
		if url.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return net.JoinHostPort(host, port)
}

func ProxyAuth(proxyURL *url.URL) string {
	if proxyURL.User == nil {
		return ""
	}
	user := proxyURL.User.Username()
	pwd, _ := proxyURL.User.Password()
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pwd))
}

// ShuffleExtensions 随机拓展重组
func ShuffleExtensions(chs *utls.ClientHelloSpec) error {
	var err error = nil
	var unshufCheck = func(idx int, exts []utls.TLSExtension) (donotshuf bool, userErr error) {
		switch exts[idx].(type) {
		case *utls.UtlsGREASEExtension:
			donotshuf = true
		case *utls.UtlsPaddingExtension, *utls.FakePreSharedKeyExtension:
			donotshuf = true
			if idx != len(chs.Extensions)-1 {
				userErr = errors.New("UtlsPaddingExtension or FakePreSharedKeyExtension must be the last extension")
			}
		default:
			donotshuf = false
		}
		return
	}
	rand.Shuffle(len(chs.Extensions), func(i, j int) {
		if unshuf, shuferr := unshufCheck(i, chs.Extensions); unshuf {
			if shuferr != nil {
				err = shuferr
			}
			return
		}
		if unshuf, shuferr := unshufCheck(j, chs.Extensions); unshuf {
			if shuferr != nil {
				err = shuferr
			}
			return
		}
		chs.Extensions[i], chs.Extensions[j] = chs.Extensions[j], chs.Extensions[i]
	})
	return err
}

// GetClientHelloSpec 按useragent获取clientHelloSpec, 默认为chrome
func GetClientHelloSpec(userAgent string) *utls.ClientHelloSpec {
	uaName, _ := useragent.New(userAgent).Browser()
	var id utls.ClientHelloID
	switch uaName {
	case "Chrome":
		id = utls.HelloChrome_Auto
	case "Firefox":
		id = utls.HelloFirefox_Auto
	case "Edge":
		id = utls.HelloEdge_Auto
	case "Safari":
		id = utls.HelloSafari_Auto
	case "360Browser":
		id = utls.Hello360_Auto
	case "QQBrowser":
		id = utls.HelloQQ_Auto
	default:
		id = utls.HelloChrome_Auto
	}
	chs, _ := utls.UTLSIdToSpec(id)
	ShuffleExtensions(&chs)
	return &chs
}
