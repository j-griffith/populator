package v1alpha1

import (
	"github.com/j-griffith/populator/pkg/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type PopulatorInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.PopulatorList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Populator, error)
	Create(*v1alpha1.Populator) (*v1alpha1.Populator, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

type populatorClient struct {
	restClient rest.Interface
	ns         string
}

func (c *populatorClient) List(opts metav1.ListOptions) (*v1alpha1.PopulatorList, error) {
	result := v1alpha1.PopulatorList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("populators").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *populatorClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Populator, error) {
	result := v1alpha1.Populator{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("populators").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *populatorClient) Create(project *v1alpha1.Populator) (*v1alpha1.Populator, error) {
	result := v1alpha1.Populator{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("populators").
		Body(project).
		Do().
		Into(&result)

	return &result, err
}

func (c *populatorClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("populators").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}
