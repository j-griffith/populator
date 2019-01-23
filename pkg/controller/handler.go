package controller

import (
	"log"

	clientset "github.com/j-griffith/populator/pkg/clientset/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// PopulatorHandler is a sample implementation of Handler
type PopulatorHandler struct {
	KubeClient      kubernetes.Interface
	PopulatorClient clientset.Interface
}

// Init handles any handler initialization
func (p *PopulatorHandler) Init() error {
	log.Println("initialize PopulatorHandler")
	return nil
}

// ObjectCreated is called when an object is created
func (p *PopulatorHandler) ObjectCreated(obj interface{}) {
	log.Println("handle ObjectCreated event")
	// assert the type to a PVC object to pull out relevant data
	pvc := obj.(*core_v1.PersistentVolumeClaim)
	pop, err := p.PopulatorClient.Populators("default").Get(pvc.Spec.DataSource.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("unable to fetch requested DataSource: %s, error: %v\n", pvc.Spec.DataSource.Name, err)
		log.Printf("PV was created but will NOT be populated\n")
		return
	}
	log.Printf("launch population process for PVC: %s, using populator: %s", pvc.Name, pop.Name)
	log.Printf("completed create event handling for pvc: %s", pvc.Name)
}

// ObjectDeleted is called when an object is deleted
func (p *PopulatorHandler) ObjectDeleted(obj interface{}) {
	log.Println("handle ObjectDeleted event")
}

// ObjectUpdated is called when an object is updated
func (p *PopulatorHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Println("handle ObjectUpdated event")
}
