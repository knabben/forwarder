package main

import (
	"fmt"
	"github.com/golang/glog"
	"time"

	"github.com/knabben/forwarder/pkg/port"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	runtime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

// NewController Start a new event listener and returns the controller
func NewController() *Controller {
	listWatcher := cache.NewListWatchFromClient(port.Clientset.CoreV1().RESTClient(),
		"pods", v1.NamespaceDefault, fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Watcher for Pod events and add it to work queue
	indexer, informer := cache.NewIndexerInformer(listWatcher, &v1.Pod{}, 0,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err == nil {
					queue.Add(key)
				}
			},
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
				}
			},
		}, cache.Indexers{})

	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

// Run starts the controller
func (c *Controller) Run(stopChannel chan struct{}) {
	glog.Info("Starting Pod controller")

	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	go c.informer.Run(stopChannel)

	if !cache.WaitForCacheSync(stopChannel, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches toz sync"))
		return
	}

	go wait.Until(c.runWorker, time.Second, stopChannel)

	<-stopChannel
	glog.Info("Stopping Pod controller.")
}

// processNextItem consume queue and execute
func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)
	_ = c.handleEvent(key.(string))

	return true
}

// runWorker keep processing events
func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) handleEvent(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Error("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if exists {
		port.StartPortForward(obj.(*v1.Pod), key)
	} else {
		port.RemovePod(key)
	}

	return nil
}
