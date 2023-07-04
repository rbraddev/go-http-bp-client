package client

import "strings"

func parseProxies(p []string) []Proxy {
	var proxies []Proxy
	for _, proxy := range p {
		res := strings.Split(proxy, ":")
		if len(res) > 2 {
			proxies = append(proxies, Proxy{host: res[0], port: res[1], username: res[2], password: res[3]})
		} else {
			proxies = append(proxies, Proxy{host: res[0], port: res[1]})
		}
	}
	return proxies
}
