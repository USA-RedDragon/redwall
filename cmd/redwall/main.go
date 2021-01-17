package main

import (
	"os"
	"path/filepath"

	"github.com/USA-RedDragon/redwall/pkg/cmd/redwall"
	"k8s.io/klog/v2"
)

func main() {
	defer klog.Flush()

	baseName := filepath.Base(os.Args[0])

	err := redwall.NewCommand(baseName).Execute()
	if err != nil {
		klog.Exit(err)
	}
}
