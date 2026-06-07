package network

import (
	"io"
	"net"
	"net/http"
)

var (
	publicIp string
	localIp  string
)

func obterIpPublico() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "0.0.0.0"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func obterIpLocal() string {
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			s := ip.String()
			if len(s) >= 7 && s[:7] == "172.16." {
				continue
			}
			return s
		}
	}
	return "127.0.0.1"
}
