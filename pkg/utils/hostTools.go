package utils

import (
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"

	"github.com/jpillora/go-tld"
)

// MatchRootDomain 匹配根域名
func MatchRootDomain(rawURL string) (string, error) {
	if !WithProtocol(rawURL) {
		rawURL = "http://" + rawURL
	}
	// 解析URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if JudgeNetProtocol(parsedURL.Host) != Invalid {
		return parsedURL.Host, nil
	}
	extracted, err := tld.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", extracted.Domain, extracted.TLD), nil
}

// WithProtocol 含协议字符串
func WithProtocol(domain string) bool {
	var match bool
	reg := "^((http://)|(https://)).+"
	match, _ = regexp.MatchString(reg, domain)
	return match
}

// IsIPv4 判断地址为IPv4
func IsIPv4(ipStr string) (flag bool) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return
	}
	return addr.Is4()
}

// IsIPv6 判断地址为IPv6
func IsIPv6(ipStr string) (flag bool) {
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return
	}
	return addr.Is6()
}

// JudgeNetProtocol 裁决ip字符串类型
func JudgeNetProtocol(ipStr string) NetProtocol {
	if IsIPv4(ipStr) {
		return IPv4
	}
	if IsIPv6(ipStr) {
		return IPv6
	}
	return Invalid
}

// TrimProtocol消除Protocol
func TrimProtocol(host string) string {
	if WithProtocol(host) {
		for _, v := range GetProtocolMap() {
			host = strings.TrimPrefix(host, v+"://")
		}
	}
	return host
}

// 设置Protocol
func SetProtocol(host string, p Protocol) string {
	host = TrimProtocol(host)
	pro, ok := GetProtocolMap()[p]
	if !ok {
		pro = GetProtocolMap()[HTTP]
	}
	return fmt.Sprintf("%s://%s", pro, host)
}
