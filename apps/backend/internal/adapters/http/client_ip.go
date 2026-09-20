package http

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type clientIPResolver func(*http.Request) string

func newClientIPResolver(trustedProxyCIDRs []netip.Prefix, edgeProxyAuthToken string) clientIPResolver {
	return func(request *http.Request) string {
		peer := remoteAddrIP(request.RemoteAddr)
		if peer.IsValid() {
			if edgeProxyTokenMatches(request.Header.Get("X-FastTourney-Edge-Token"), edgeProxyAuthToken) {
				if forwarded, err := netip.ParseAddr(strings.TrimSpace(request.Header.Get("X-Client-IP"))); err == nil {
					return forwarded.Unmap().String()
				}
			}
			for _, trustedProxyCIDR := range trustedProxyCIDRs {
				if trustedProxyCIDR.Contains(peer) {
					if forwarded, err := netip.ParseAddr(strings.TrimSpace(request.Header.Get("X-Client-IP"))); err == nil {
						return forwarded.Unmap().String()
					}
					break
				}
			}
			return peer.String()
		}
		return request.RemoteAddr
	}
}

func edgeProxyTokenMatches(provided, expected string) bool {
	if expected == "" || len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func remoteAddrIP(remoteAddr string) netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return address.Unmap()
}
