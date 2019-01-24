package controller

import (
	"log"

	clientset "github.com/j-griffith/populator/pkg/clientset/v1alpha1"
	"github.com/j-griffith/populator/pkg/populator"
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

	// If there's no DS specified just ignore it and move along (and don't vomit when you try and acess the field)
	if pvc.Spec.DataSource == nil {
		log.Printf("no DataSource entry for PVC %s, moving along", pvc.Name)
		return
	}

	// TODO: throw in some error checking so we don't hit nil pointer type crashes if somebody didn't fill this out correctly
	// Some of it we handle with the requirements in the CRD, others we can add webhooks, but for now living on the edge
	pop, err := p.PopulatorClient.Populators("default").Get(pvc.Spec.DataSource.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("unable to fetch requested DataSource: %s, error: %v\n", pvc.Spec.DataSource.Name, err)
		log.Printf("PV was created but will NOT be populated\n")
		return
	}
	// CreateJobFromObjects creates the job spec and launches it
	job, err := populator.CreateJobFromObjects(p.KubeClient, pvc, pop)
	if err != nil {

	}
	log.Printf("succesfully launch a populator job (%v) for PVC %s", job, pvc.Name)
	// TODO
	// * We'll need to add some tracking type data to the PVC in the form of attributes
}

// ObjectDeleted is called when an object is deleted
func (p *PopulatorHandler) ObjectDeleted(obj interface{}) {
	log.Println("handle ObjectDeleted event")
}

// ObjectUpdated is called when an object is updated
func (p *PopulatorHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Println("handle ObjectUpdated event")
}
