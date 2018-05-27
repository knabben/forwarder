package main

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	home = os.Getenv("HOME")

	forwardFile string
	kubeConfig  string

	rootCmd = &cobra.Command{
		Use:   "forwarder",
		Short: "Forward",
		Long:  "Kubernetes local service port forward.",
		RunE: func(c *cobra.Command, args []string) error {
			initLogs()

			kubernetesSetup()

			// Start controller and run
			controller := NewController()

			stop := make(chan struct{})
			go controller.Run(stop)

			WaitSignal(stop)
			glog.Flush()

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig",
		filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&forwardFile, "file",
		filepath.Join(home, ".kube", "forwarders.json"), "forwarders JSON file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
		os.Exit(1)
	}
}
