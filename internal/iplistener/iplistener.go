package iplistener

import (
	"fmt"
	"net"
	"time"

	"github.com/asaskevich/EventBus"
	"k8s.io/klog/v2"
)

type IPListener struct {
	currentIP net.IP
	bus       EventBus.Bus
	ipv6      bool
	busTopic  string
}

// New will create the IP address change listener
func New(bus EventBus.Bus, currentIP net.IP, ipv6 bool) *IPListener {
	var ipType string
	if ipv6 {
		ipType = "ipv6"
	} else {
		ipType = "ipv4"
	}
	topic := fmt.Sprintf("%s:changed", ipType)
	return &IPListener{
		currentIP,
		bus,
		ipv6,
		topic,
	}
}

func (l IPListener) Subscribe(fn interface{}) {
	err := l.bus.Subscribe(l.busTopic, fn)
	if err != nil {
		klog.Error(err)
	}
}

func (l IPListener) Start() {
	for {
		ipv4, ipv6, err := GetInterfaceIPAddr("wan")
		if err != nil {
			klog.Error(err)
		}

		var ip net.IP

		if l.ipv6 {
			ip = ipv6
		} else {
			ip = ipv4
		}

		if ip == nil && l.currentIP != nil {
			klog.Info("Publishing empty IP")
			l.currentIP = ip
			l.bus.Publish(l.busTopic, ip)
			continue
		} else if ip != nil && !ip.Equal(l.currentIP) {
			klog.Infof("IP changed to %v", ip)
			l.currentIP = ip
			l.bus.Publish(l.busTopic, ip)
		}

		// Just to avoid constant spamming of interface checks
		time.Sleep(time.Millisecond * 250)
	}
}
