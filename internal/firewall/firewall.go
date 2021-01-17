package firewall

import (
	"fmt"
	"net"
	"os"

	"os/exec"

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
			f, err := os.OpenFile("/etc/nftables/public_ip.nft", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				klog.Fatal(err)
			}
			_, err = f.WriteString(fmt.Sprintf("define public_ip = %s\n", newIP.String()))
			if err != nil {
				klog.Fatal(err)
			}
			if err := f.Close(); err != nil {
				klog.Fatal(err)
			}
			cmd := exec.Command("nft", "-f", "/etc/nftables.conf")
			err = cmd.Run()
			if err != nil {
				klog.Errorf("Failed to setup firewall: %v", err)
			}
		}
	}
}

func (f *Firewall) Start() {
	klog.Info("Configuring firewall")
	f.updateNATReflection(f.current4IP)
	f.ip4listener.Subscribe(f.updateNATReflection)

	cmd := exec.Command("nft", "-f", "/etc/nftables.conf")
	err := cmd.Run()
	if err != nil {
		klog.Errorf("Failed to setup firewall: %v", err)
	}

	klog.Info("Firewall setup")
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
