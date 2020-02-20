package main

import (
	fmt "fmt"

	"github.com/dgraph-io/ristretto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesClient struct {
	dynamic.Interface
	meta.RESTMapper
	*kubernetes.Clientset
}

func buildKubernetesConfig(kubeconfig string) (config *rest.Config, err error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
}

func newK8sClient(kubeconf string) (*kubernetesClient, error) {
	config, err := buildKubernetesConfig(kubeconf)
	if err != nil {
		return nil, err
	}

	intf, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clientset := kubernetes.NewForConfigOrDie(config)

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		Clientset:  clientset,
		Interface:  intf,
		RESTMapper: restmapper.NewDiscoveryRESTMapper(groupResources),
	}, nil
}

func (kc *kubernetesClient) getNodeAddress(cache *ristretto.Cache, node string) ([]string, error) {

	if node == "" {
		return []string{}, nil
	}

	ident := fmt.Sprintf("node-%v", node)
	cached, ok := cache.Get(ident)
	if ok {
		if t, ok := cached.([]string); ok {
			return t, nil
		}
	}

	n, err := kc.Clientset.CoreV1().Nodes().Get(node, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	address := []string{}
	for _, i := range n.Status.Addresses {
		address = append(address, i.Address)
	}

	defer cache.Set(ident, address, 1)
	return address, nil
}

func (kc *kubernetesClient) getObject(cache *ristretto.Cache, ref *v1.ObjectReference) (*unstructured.Unstructured, error) {
	uid := string(ref.UID)

	cached, ok := cache.Get(uid)
	if ok {
		return cached.(*unstructured.Unstructured), nil
	}

	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}

	gk := schema.GroupKind{Group: gv.Group, Kind: ref.Kind}
	mapping, err := kc.RESTMapper.RESTMapping(gk, gv.Version)
	if err != nil {
		return nil, err
	}

	item, err := kc.Interface.Resource(mapping.Resource).Namespace(ref.Namespace).Get(ref.Name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	defer cache.Set(uid, item, 1)
	if ref.UID != item.GetUID() {
		defer cache.Set(string(item.GetUID()), item, 1)
	}

	return item, nil
}
