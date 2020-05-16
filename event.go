package main

import (
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
)

type L9Event struct {
	ID                 string                 `json:"id"`
	Timestamp          int64                  `json:"timestamp"`
	Component          string                 `json:"component"`
	Host               string                 `json:"host"`
	Message            string                 `json:"message"`
	Namespace          string                 `json:"namespace"`
	Reason             string                 `json:"reason"`
	ReferenceUID       string                 `json:"reference_uid"`
	ReferenceNamespace string                 `json:"reference_namespace"`
	ReferenceName      string                 `json:"reference_name"`
	ReferenceKind      string                 `json:"reference_kind"`
	ReferenceVersion   string                 `json:"reference_version"`
	ObjectUid          string                 `json:"object_uid"`
	Labels             map[string]string      `json:"labels"`
	Annotations        map[string]string      `json:"annotations"`
	Address            []string               `json:"address"`
	Pod                map[string]interface{} `json:"pod"`
	Version            string                 `json:"version"`
}

func makeL9Event(
	db Cachier, c *kubernetesClient, e *v1.Event,
) (*L9Event, error) {
	u, err := c.getObject(db, &e.InvolvedObject)
	if err != nil {
		return nil, err
	}

	address, err := c.getNodeAddress(db, e.Source.Host)
	if err != nil {
		return nil, err
	}

	return makeL9EventDetails(db, e, u, address)
}

func makeL9EventDetails(db Cachier, e *v1.Event, u *unstructured.Unstructured, address []string) (*L9Event, error) {
	ne := &L9Event{
		ID:                 string(e.UID),
		Timestamp:          e.CreationTimestamp.Time.Unix(),
		Component:          e.Source.Component,
		Host:               e.Source.Host,
		Message:            e.Message,
		Namespace:          e.Namespace,
		Reason:             e.Reason,
		ReferenceUID:       string(e.InvolvedObject.UID),
		ReferenceName:      e.InvolvedObject.Name,
		ReferenceVersion:   e.InvolvedObject.APIVersion,
		ReferenceNamespace: e.InvolvedObject.Namespace,
		ReferenceKind:      e.InvolvedObject.Kind,
		ObjectUid:          string(e.InvolvedObject.UID),
		Address:            address,
		Version:            VERSION,
	}

	if u != nil {
		ne.Labels = u.GetLabels()
		ne.Annotations = u.GetAnnotations()
		if err := addPodDetails(db, ne, u); err != nil {
			log.Println(err)
		}
	}

	return ne, nil
}

func addPodDetails(db Cachier, ne *L9Event, u *unstructured.Unstructured) error {
	p, err := unstructuredToPod(u)
	if err != nil {
		return err
	}

	ne.Pod = miniPodInfo(*p)
	return err
}

func miniPodInfo(p v1.Pod) map[string]interface{} {
	ne := map[string]interface{}{}
	ne["uid"] = p.GetUID()
	ne["name"] = p.GetName()
	ne["namespace"] = p.GetNamespace()
	ne["start_time"] = p.Status.StartTime
	ne["ip"] = p.Status.PodIP
	ne["host_ip"] = p.Status.HostIP
	return ne
}

func unstructuredToPod(obj *unstructured.Unstructured) (*v1.Pod, error) {
	json, err := runtime.Encode(unstructured.UnstructuredJSONScheme, obj)
	if err != nil {
		return nil, err
	}

	pod := new(v1.Pod)
	err = runtime.DecodeInto(clientscheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), json, pod)
	pod.Kind = ""
	pod.APIVersion = ""
	return pod, err
}
