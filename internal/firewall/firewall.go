package firewall

import (
	"net"

	"github.com/USA-RedDragon/redwall/internal/iplistener"
	"k8s.io/klog/v2"
)

type Firewall struct {
	current4IP  net.IP
	ip4listener *iplistener.IPListener
}

func (f *Firewall) updateNATReflection(newIP net.IP) {
	if newIP != nil {
		klog.Infof("Received IPv4 update: %v", newIP)

		if !newIP.Equal(f.current4IP) {
			f.current4IP = newIP
			klog.Infof("IP Changed. Updating NAT4 Reflection Rule")
			// Update reflection
			// iifname lan ip daddr $public_ip tcp dport {80, 443} dnat $nginx_server
		}
	}
}

func (f *Firewall) Start() {
	klog.Info("Configuring firewall")

	// Run nft on host with templated file

	klog.Info("Firewall setup")

	f.updateNATReflection(f.current4IP)
	f.ip4listener.Subscribe(f.updateNATReflection)
}

func New(current4IP net.IP, ip4listener *iplistener.IPListener) *Firewall {
	if current4IP == nil {
		klog.Warning("Firewall doesn't have an IP for NAT reflection")
	}

	return &Firewall{
		current4IP,
		ip4listener,
	}
}
