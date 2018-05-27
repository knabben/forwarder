package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/knabben/forwarder/pkg/port"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/golang/glog"
)

func initLogs() {
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
}

// kubernetesSetup populates the Kubernetes configuration structs
func kubernetesSetup() {
	var err error

	port.Config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	port.Clientset, err = kubernetes.NewForConfig(port.Config)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
}

// WaitSignal awaits for SIGINT or SIGTERM and closes the channel
func WaitSignal(stop chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	glog.Warningln("Finishing with signal handling.")
	close(stop)
}
