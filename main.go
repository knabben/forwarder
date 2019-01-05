package main

import (
	"os"
	"github.com/getsentry/raven-go"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	home = os.Getenv("HOME")
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
	raven.SetDSN(os.Getenv("RAVEN_DSN"))

	rootCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
		os.Exit(1)
	}
}
