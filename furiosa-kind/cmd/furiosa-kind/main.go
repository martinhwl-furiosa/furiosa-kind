/*
https://github.com/klueska/nvkind
*/

package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"k8s.io/klog/v2"
)

// Version is the version of this CLI (overwritable at build time)
// go build -ldflags "-X main.Version=1.0.0" -o furiosa-kind
var Version = "devel"

func main() {
	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "furiosa-kind"
	c.Usage = "kind for use with Furiosa NPUs"
	c.Version = Version
	c.EnableBashCompletion = true

	// Register the subcommands with the top-level CLI
	c.Commands = []*cli.Command{
		BuildClusterCommand(),
	}

	// Run the CLI
	err := c.Run(os.Args)
	if err != nil {
		klog.Fatalf("Error: %v", err)
	}
}
