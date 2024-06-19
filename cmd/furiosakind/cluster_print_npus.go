package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/martinhwl-furiosa/furiosa-kind/pkg/furiosakind"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type NodeGPUs struct {
	Node    string                `json:"node"`
	GPUInfo []furiosakind.GPUInfo `json:"gpus"`
}

type ClusterPrintGPUsFlags struct {
	Name       string
	KubeConfig string
}

func BuildClusterPrintGPUsCommand() *cli.Command {
	flags := ClusterPrintGPUsFlags{}

	cmd := cli.Command{}
	cmd.Name = "print-gpus"
	cmd.Usage = "print all NVIDIA GPUs available in a cluster"
	cmd.Action = func(ctx *cli.Context) error {
		return runClusterPrintGPUs(ctx, &flags)
	}

	cmd.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "the name of the cluster to print GPUs for",
			Destination: &flags.Name,
			EnvVars:     []string{"KIND_CLUSTER_NAME"},
		},
		&cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "Absolute path to the `KUBECONFIG` file. Either this flag or the KUBECONFIG env variable need to be set if the driver is being run out of cluster.",
			Destination: &flags.KubeConfig,
			EnvVars:     []string{"KUBECONFIG"},
		},
	}

	return &cmd
}

func runClusterPrintGPUs(c *cli.Context, f *ClusterPrintGPUsFlags) error {
	if err := f.updateFlagsWithDefaults(); err != nil {
		return fmt.Errorf("updating flags with defaults: %w", err)
	}

	clusters, err := furiosakind.GetClusterNames()
	if err != nil {
		return fmt.Errorf("getting cluster names: %w", err)
	}

	if !clusters.Has(f.Name) {
		return fmt.Errorf("unknown cluster: %v", f.Name)
	}

	cluster, err := furiosakind.NewCluster(furiosakind.WithName(f.Name))
	if err != nil {
		return fmt.Errorf("getting cluster: %w", err)
	}

	nodes, err := cluster.GetNodes()
	if err != nil {
		return fmt.Errorf("getting nodes: %w", err)
	}

	var nodeGPUsList []NodeGPUs
	for _, node := range nodes {
		if !node.HasGPUs() {
			continue
		}
		gpuInfo, err := node.GetGPUInfo()
		if err != nil {
			return fmt.Errorf("getting GPU info on node '%v': %w", node.Name, err)
		}
		nodeGPUs := NodeGPUs{
			Node:    node.Name,
			GPUInfo: gpuInfo,
		}
		nodeGPUsList = append(nodeGPUsList, nodeGPUs)
	}

	jsonData, err := json.MarshalIndent(nodeGPUsList, "", "    ")
	if err != nil {
		return fmt.Errorf("marshaling GPU info: %w", err)
	}
	fmt.Println(string(jsonData))

	return nil
}

func (f *ClusterPrintGPUsFlags) updateFlagsWithDefaults() error {
	if f.KubeConfig == "" {
		if home := homedir.HomeDir(); home != "" {
			f.KubeConfig = home + "/.kube/config"
		}
	}

	if f.Name != "" {
		return nil
	}

	config, err := clientcmd.LoadFromFile(f.KubeConfig)
	if err != nil {
		return fmt.Errorf("marshaling GPU info: %w", err)
	}

	if config.CurrentContext == "" {
		return fmt.Errorf("no current kubecontext set")
	}

	if !strings.HasPrefix(config.CurrentContext, "kind-") {
		return fmt.Errorf("current kubecontext is not a kind cluster: %v", config.CurrentContext)
	}

	f.Name = strings.TrimPrefix(config.CurrentContext, "kind-")

	return nil
}
