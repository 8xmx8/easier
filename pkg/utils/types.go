package utils

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// NetProtocol
type NetProtocol int

const (
	Invalid NetProtocol = iota // 0. 无法解析
	IPv4                       // 1. IPv4
	IPv6                       // 2.IPv6
)

func (np NetProtocol) String() string {
	switch np {
	case IPv4:
		return "IPv4"
	case IPv6:
		return "IPv6"
	// case Invalid:
	// 	return "Invalid"
	default:
		return "Invalid"
	}
}

// 域名解析类型
type ResolveType string

const (
	DefaultType ResolveType = ""     // 默认空值
	Atype       ResolveType = "A"    // IPv4
	AAAAtype    ResolveType = "AAAA" // IPv6
	CNAMEtype   ResolveType = "CNAME"
	MXtype      ResolveType = "MX"
	NStype      ResolveType = "NS"
	PTRtype     ResolveType = "PTR"
	SOAtype     ResolveType = "SOA"
)

// Protocol 网络通信协议
type Protocol int

const (
	HTTP Protocol = iota + 1
	HTTPS
)

func (p Protocol) String() string {
	switch p {
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	}
	return ""
}

func GetProtocolMap() map[Protocol]string {
	return map[Protocol]string{
		HTTP:  "http",
		HTTPS: "https",
	}
}
