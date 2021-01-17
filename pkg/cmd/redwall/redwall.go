package redwall

import (
	"flag"

	"github.com/USA-RedDragon/redwall/pkg/cmd/start"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	Ddns bool
)

func NewCommand(name string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: "Firewall and DDNS client based on nftables",
		Long:  `TODO: Long description`,
	}

	c.AddCommand(
		start.NewCommand(),
	)

	// init and add the klog flags
	klog.InitFlags(flag.CommandLine)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return c
}
