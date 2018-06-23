package port

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fmt"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	kb "k8s.io/client-go/kubernetes"
	rc "k8s.io/client-go/rest"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Pod struct {
	Name      string
	Label     string
	Port      string
	CloseChan chan struct{}
	ReadyChan chan struct{}
}

var (
	Pods map[string]Pod = make(map[string]Pod)

	Clientset *kb.Clientset
	Config    *rc.Config
)

// StartPortForward add a new pod and start port binding
func StartPortForward(obj *v1.Pod, key string) error {
	firstRunningState := false

	label := obj.Labels["app"]
	phase := obj.Status.Phase

	name := obj.GetName()
	deletionTimestamp := obj.DeletionTimestamp

	for _, port := range fetchServicePort(label) {
		port := strconv.Itoa(int(port))
		AddPod(key, name, label, port)
	}

	// Check pod condition
	for _, item := range obj.Status.Conditions {
		if phase == v1.PodRunning && item.Type == v1.PodReady &&
			item.Status == v1.ConditionTrue && deletionTimestamp == nil {

			firstRunningState = true
		}
	}

	if firstRunningState {
		for _, port := range fetchServicePort(label) {
			port := strconv.Itoa(int(port))
			podKey := CreateKey(key, port)

			go func(pod Pod) {
				defer func() {
					glog.Warningln("So long and thanks for all the fish")
					glog.Flush()
				}()

				port := pod.Port

				glog.Warningln("Listening", pod.Name, "gorouting start on", port)

				req := Clientset.CoreV1().RESTClient().Post().Resource("pods").
					Namespace("default").Name(pod.Name).SubResource("portforward")
				ForwardPort(
					"POST",
					req.URL(),
					Config,
					[]string{port},
					pod.CloseChan,
					pod.ReadyChan,
				)
			}(Pods[podKey])
		}
	}
	return nil
}

func CreateKey(key, port string) string {
	return fmt.Sprintf("%s-%s", key, port)
}

//AddPod on running list
func AddPod(key string, name string, label string, port string) {
	podKey := CreateKey(key, port)

	if _, ok := Pods[podKey]; !ok {
		closeChan := make(chan struct{})
		readyChan := make(chan struct{})

		glog.Infoln("-> Starting pod: ", name, key)

		Pods[podKey] = Pod{
			Name:      name,
			Port:      port,
			CloseChan: closeChan,
			ReadyChan: readyChan,
			Label:     label,
		}
	}
}

// RemovePod from running list and close communication channel
func RemovePod(key string) {
	for podKey := range Pods {
		if strings.Contains(podKey, key) {
			pod := Pods[podKey]

			close(pod.CloseChan)
			delete(Pods, podKey)

			glog.Warningln(key, " listening closed: ", pod.Port)
		}
	}
}

// fetchServicePort fetches service port with label
func fetchServicePort(label string) []int32 {
	var ports []int32

	req, _ := Clientset.CoreV1().Services("default").List(meta.ListOptions{
		LabelSelector: "app=" + label,
	})
	for _, service := range req.Items {
		for _, port := range service.Spec.Ports {
			ports = append(ports, port.Port)
		}
	}
	return ports
}

// StartForwardingPorts initialize port forwarding based on json file
func ForwardPort(method string, url *url.URL, config *rc.Config, ports []string, stopChannel chan struct{}, readyChannel chan struct{}) error {
	var cmdOut, cmdErr io.Writer

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		glog.Error(err)
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.New(dialer, ports, stopChannel, readyChannel, cmdOut, cmdErr)
	if err != nil {
		glog.Error("NEW PORT ", err)
		return err
	}

	err = fw.ForwardPorts()
	if err != nil {
		glog.Error("PORT FORWARD ", err)
		return err
	}

	return nil
}
