package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/j-griffith/populator/pkg/api/types/v1alpha1"
	clientV1alpha1 "github.com/j-griffith/populator/pkg/clientset/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig string

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to Kubernetes config file")
	flag.Parse()
}

func main() {
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

	v1alpha1.AddToScheme(scheme.Scheme)

	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	/*
		projects, err := clientSet.Populators("default").List(metav1.ListOptions{})
		if err != nil {
			panic(err)
		}
	*/
	p, err := clientSet.Populators("default").Get("my-populator", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\nPopulator:\n%+v\n", p)

}