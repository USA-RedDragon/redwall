package iplistener

import (
	"net"
)

func GetInterfaceIPAddr(interfaceName string) (addr4 net.IP, addr6 net.IP, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
		ipv6Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil {
		return nil, nil, err
	}
	if addrs, err = ief.Addrs(); err != nil {
		return nil, nil, err
	}
	for _, addr := range addrs {
		if ipv4Addr != nil && ipv6Addr != nil {
			break
		}
		v4 := addr.(*net.IPNet).IP.To4()
		var v6 net.IP
		if v4 == nil {
			v6 = addr.(*net.IPNet).IP.To16()
			if v6 != nil && ipv6Addr == nil {
				ipv6Addr = v6
			}
		} else if ipv4Addr == nil {
			ipv4Addr = v4
		}
	}
	return ipv4Addr, ipv6Addr, nil
}
