package network

import (
	"net"
)

const DefaultSVRIP = "127.0.0.1"

func GetSVRIP(deviceName string) string {
	if deviceName == "" {
		return DefaultSVRIP
	}
	if ip := getSVRIP(deviceName); ip != "" {
		return ip
	}
	return DefaultSVRIP
}

func getSVRIP(deviceName string) string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifs {
		if iface.Name == deviceName && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}
