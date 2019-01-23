package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"log"

	"github.com/j-griffith/populator/pkg/clientset/v1alpha1"
	ctrl "github.com/j-griffith/populator/pkg/controller"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	papi "github.com/j-griffith/populator/pkg/api/types/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

var kubeconfig string

/*
// getKubeConfig fetches our kubeconfig, we're not really doing anything here, if you passed a kubeconfig path in
// when running we'll attempt to use that, otherwise we'll assume running in-cluster and just leverage teh InClusterConfig
func getKubeConfig() *rest.Config {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		log.Printf("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err)
	}
	return config
}
*/

// getPopulatorClient sets up a Populator client using the provided config
func getPopulatorClient(cfg *rest.Config) v1alpha1.Interface {
	clientSet, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	return clientSet
}

// getKubernetesClient sets up a K8s client using the provided config
func getKubernetesClient(cfg *rest.Config) kubernetes.Interface {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("get kubernetes client: %v", err)
	}

	log.Println("Successfully constructed k8s client")
	return client
}

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to Kubernetes config file")
	flag.Parse()

}

// main code path
func main() {
	var cfg *rest.Config
	var err error

	if kubeconfig == "" {
		log.Println("using in-cluster kube config")
		cfg, err = rest.InClusterConfig()
	} else {
		log.Printf("using supplied kube config (%s)", kubeconfig)
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	k8sClient := getKubernetesClient(cfg)
	populatorClient := getPopulatorClient(cfg)
	papi.AddToScheme(scheme.Scheme)

	/*
		// get the Kubernetes client for connectivity
		client, cfg := getKubernetesClient()
	*/

	// create the informer so that we can not only list resources
	// but also watch them for all PVCs in the default namespace
	informer := cache.NewSharedIndexInformer(
		// the ListWatch contains two different functions that our
		// informer requires: ListFunc to take care of listing and watching
		// the resources we want to handle
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				// list all of the pvcs (core resource) in the deafult namespace
				return k8sClient.CoreV1().PersistentVolumeClaims(metav1.NamespaceDefault).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				// watch all of the pvcs (core resource) in the default namespace
				return k8sClient.CoreV1().PersistentVolumeClaims(metav1.NamespaceDefault).Watch(options)
			},
		},
		&api_v1.PersistentVolumeClaim{}, // the target type (PVC)
		0,                               // no resync (period of 0)
		cache.Indexers{},
	)

	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Printf("Add PVC: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// We only care about DataSource updates in this controller, so filter out anything that's not updating/adding a DataSource entry and move along
			origPVC, _ := oldObj.(*api_v1.PersistentVolumeClaim)
			updatedPVC, _ := newObj.(*api_v1.PersistentVolumeClaim)
			if origPVC.Spec.DataSource != updatedPVC.Spec.DataSource {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				log.Printf("Update PVC: %s", key)
				if err == nil {
					queue.Add(key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Printf("Delete PVC: %s", key)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	// construct the Controller object which has all of the necessary components to
	// handle logging, connections, informing (listing and watching), the queue,
	// and the handler
	controller := ctrl.Controller{
		ClientSet:          k8sClient,
		Informer:           informer,
		Queue:              queue,
		Handler:            &ctrl.PopulatorHandler{KubeClient: k8sClient, PopulatorClient: populatorClient},
		PopulatorClientSet: populatorClient,
	}

	// use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the controller loop to process items
	go controller.Run(stopCh)

	// use a channel to handle OS signals to terminate and gracefully shut
	// down processing
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
