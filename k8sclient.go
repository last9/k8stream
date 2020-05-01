package main

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	objectCacheTable  = "object"
	objectCacheExpiry = 600
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

func (kc *kubernetesClient) getApps(db Cachier, s *v1.Service) ([]appsv1.Deployment, error) {
	namespace := s.GetNamespace()

	q := labels.Set(s.Spec.Selector)
	apps, err := kc.Clientset.AppsV1().Deployments(namespace).List(
		metav1.ListOptions{LabelSelector: q.String()},
	)
	if err != nil {
		return nil, err
	}

	return apps.Items, nil
}

func (kc *kubernetesClient) getPods(db Cachier, s *v1.Service) ([]v1.Pod, error) {
	namespace := s.GetNamespace()
	q := labels.Set(s.Spec.Selector)
	pods, err := kc.Clientset.CoreV1().Pods(namespace).List(
		metav1.ListOptions{LabelSelector: q.String()},
	)

	if err != nil {
		return nil, err
	}

	return pods.Items, nil

}

func (kc *kubernetesClient) getService(namespace, name string) (*v1.Service, error) {
	return kc.Clientset.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

func (kc *kubernetesClient) getNodeAddress(db Cachier, node string) ([]string, error) {
	addr := []string{}

	if node == "" {
		return addr, nil
	}

	res, err := db.Get("node", node)
	if err != nil {
		return nil, err
	}

	if res.Exists() {
		return addr, res.Unmarshal(&addr)
	}

	n, err := kc.Clientset.CoreV1().Nodes().Get(node, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, i := range n.Status.Addresses {
		addr = append(addr, i.Address)
	}

	defer db.ExpireSet("node", node, addr, objectCacheExpiry)
	return addr, nil
}

func (kc *kubernetesClient) getObject(db Cachier, ref *v1.ObjectReference) (*unstructured.Unstructured, error) {
	uid := string(ref.UID)

	var cached *unstructured.Unstructured
	result, err := db.Get(objectCacheTable, uid)
	if err != nil {
		return nil, err
	}

	if result.Exists() {
		return cached, result.Unmarshal(&cached)
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

	defer db.ExpireSet(objectCacheTable, uid, item, objectCacheExpiry)
	if ref.UID != item.GetUID() {
		defer db.ExpireSet(objectCacheTable, string(item.GetUID()), item, objectCacheExpiry)
	}

	return item, nil
}
