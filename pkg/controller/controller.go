package controller

import (
	"fmt"
	"time"

	"log"

	clientset "github.com/j-griffith/populator/pkg/clientset/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var myKubeClient kubernetes.Interface

// Controller struct defines how a controller should encapsulate
// client connectivity, informing (list and watching)
// queueing, and handling of resource changes
type Controller struct {
	ClientSet          kubernetes.Interface
	Queue              workqueue.RateLimitingInterface
	Informer           cache.SharedIndexInformer
	Handler            Handler
	PopulatorClientSet clientset.Interface
}

// Run is the main path of execution for the controller loop
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.Queue.ShutDown()

	log.Printf("starting the populator-controller")

	// run the informer to start listing and watching resources
	go c.Informer.Run(stopCh)

	// do the initial synchronization (one time) to populate resources
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	log.Printf("cache sync complete")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced allows us to satisfy the Controller interface
// by wiring up the informer's HasSynced method to it
func (c *Controller) HasSynced() bool {
	return c.Informer.HasSynced()
}

// runWorker executes the loop to process new items added to the queue
func (c *Controller) runWorker() {
	log.Printf("starting the populator-controller worker thread")

	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource
	for c.processNextItem() {
		log.Printf("populator-controller processing next item")
	}
	log.Printf("populator-controller finished processing item")
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *Controller) processNextItem() bool {
	log.Println("populator-controller process next item begin")

	// fetch the next item (blocking) from the queue to process or
	// if a shutdown is requested then return out of this to stop
	// processing
	key, quit := c.Queue.Get()

	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the queue has indicated
	// from the Get method
	if quit {
		return false
	}

	defer c.Queue.Done(key)

	// assert the string out of the key (format `namespace/name`)
	keyRaw := key.(string)

	// take the string key and get the object out of the indexer
	//
	// item will contain the complex object for the resource and
	// exists is a bool that'll indicate whether or not the
	// resource was created (true) or deleted (false)
	//
	// if there is an error in getting the key from the index
	// then we want to retry this particular queue key a certain
	// number of times (5 here) before we forget the queue key
	// and throw an error
	item, exists, err := c.Informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.Queue.NumRequeues(key) < 5 {
			log.Printf("failed processing item with key %s, error: %v (attempting retries)", key, err)
			c.Queue.AddRateLimited(key)
		} else {
			log.Printf("failed procesing item with key %s, error: %v (no retries left)", key, err)
			c.Queue.Forget(key)
			utilruntime.HandleError(err)
		}
	}

	// if the item doesn't exist then it was deleted and we need to fire off the handler's
	// ObjectDeleted method. but if the object does exist that indicates that the object
	// was created (or updated) so run the ObjectCreated method
	//
	// after both instances, we want to forget the key from the queue, as this indicates
	// a code path of successful queue key processing
	if !exists {
		log.Printf("object delete detected: %s", keyRaw)
		c.Handler.ObjectDeleted(item)
		c.Queue.Forget(key)
	} else {
		log.Printf("object create detected: %s", keyRaw)
		c.Handler.ObjectCreated(item)
		c.Queue.Forget(key)
	}

	// keep the worker loop running by returning true
	return true
}
