package start

import (
	"github.com/USA-RedDragon/redwall/internal/ddns/cloudflare"
	"github.com/USA-RedDragon/redwall/internal/firewall"
	"github.com/USA-RedDragon/redwall/internal/iplistener"
	"github.com/asaskevich/EventBus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "start [OPTIONS]",
		Short: "Start the redwall daemon",
		Run: func(c *cobra.Command, args []string) {
			var ip4listener *iplistener.IPListener
			var ip6listener *iplistener.IPListener

			var ip4CfDDNS *cloudflare.CloudflareDDNS
			var ip6CfDDNS *cloudflare.CloudflareDDNS

			var redwall *firewall.Firewall

			eventBus := EventBus.New()

			ddnsEnabled, err := c.Flags().GetBool("ddns")
			if err != nil {
				klog.Fatal(err)
				return
			}

			ipv4, ipv6, err := iplistener.GetInterfaceIPAddr("wan")
			if err != nil {
				klog.Error(err)
			}

			klog.Infof("ipv4addr = %v\n", ipv4)
			klog.Infof("ipv6addr = %v\n", ipv6)

			ip4listener = iplistener.New(eventBus, ipv4, false)
			go ip4listener.Start()

			ip6listener = iplistener.New(eventBus, ipv6, true)
			go ip6listener.Start()

			redwall = firewall.New(ipv4, ip4listener)
			go redwall.Start()

			if ddnsEnabled {
				ip4CfDDNS = cloudflare.New(ipv4, false, ip4listener)
				if ip4CfDDNS != nil {
					go ip4CfDDNS.Start()
				} else {
					klog.Warning("IPv4 Cloudflare DDNS failed to start")
				}

				ip6CfDDNS = cloudflare.New(ipv6, true, ip6listener)
				if ip6CfDDNS != nil {
					go ip6CfDDNS.Start()
				} else {
					klog.Warning("IPv6 Cloudflare DDNS failed to start")
				}
			}

			for {
				c := make(chan int)
				<-c
			}
		},
	}

	c.PersistentFlags().Bool("ddns", false, "Whether to enable the DDNS service")

	return c
}
