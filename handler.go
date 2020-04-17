package main

import (
	"log"
	"strings"

	"github.com/dgraph-io/ristretto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
)

type Handler struct {
	client *kubernetesClient
	ch     chan<- interface{}
	*ristretto.Cache
}

func (h *Handler) OnAdd(obj interface{}) {
	event := obj.(*v1.Event)
	h.onEvent(event)
}

func (h *Handler) OnUpdate(oldObj, newObj interface{}) {
	event := newObj.(*v1.Event)
	h.onEvent(event)
}

func (h *Handler) OnDelete(obj interface{}) {
	// Ignore deletes
	event := obj.(*v1.Event)
	h.onEvent(event)
}

func (h *Handler) onEvent(e *v1.Event) {
	if _, ok := h.Cache.Get(string(e.UID)); ok {
		return
	}

	obj, err := h.client.getObject(h.Cache, &e.InvolvedObject)
	if err != nil {
		log.Println(err)
	}

	addr, err := h.client.getNodeAddress(h.Cache, e.Source.Host)
	if err != nil {
		log.Println(err)
	}

	h.ch <- makeL9Event(e, obj, addr)
}

type L9Event struct {
	ID                 string                 `json:"id,omitempty"`
	Timestamp          int64                  `json:"timestamp,omitempty"`
	Component          string                 `json:"component,omitempty"`
	Host               string                 `json:"host,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Namespace          string                 `json:"namespace,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	ReferenceUID       string                 `json:"reference_uid,omitempty"`
	ReferenceNamespace string                 `json:"reference_namespace,omitempty"`
	ReferenceName      string                 `json:"reference_name,omitempty"`
	ReferenceKind      string                 `json:"reference_kind,omitempty"`
	ReferenceVersion   string                 `json:"reference_version,omitempty"`
	ObjectUid          string                 `json:"object_uid,omitempty"`
	Labels             map[string]string      `json:"labels,omitempty"`
	Annotations        map[string]string      `json:"annotations,omitempty"`
	Address            []string               `json:"address,omitempty"`
	Pod                map[string]interface{} `json:"pod,omitempty"`
}

func makeL9Event(
	e *v1.Event, u *unstructured.Unstructured, address []string,
) *L9Event {
	ne := &L9Event{
		ID:               string(e.UID),
		Timestamp:        e.CreationTimestamp.Time.Unix(),
		Component:        e.Source.Component,
		Host:             e.Source.Host,
		Message:          e.Message,
		Namespace:        e.Namespace,
		Reason:           e.Reason,
		ReferenceUID:     string(e.InvolvedObject.UID),
		ReferenceName:    e.InvolvedObject.Name,
		ReferenceVersion: e.InvolvedObject.APIVersion,
		Address:          address,
	}

	if u != nil {
		if err := addPodDetails(ne, u); err != nil {
			log.Println(err)
		}

		// ne.InvolvedObject = u
		ne.ReferenceNamespace = u.GetNamespace()
		ne.ReferenceKind = u.GetKind()
		ne.ObjectUid = string(u.GetUID())
		ne.Labels = u.GetLabels()
		ne.Annotations = u.GetAnnotations()
	}

	return ne
}

func addPodDetails(ne *L9Event, u *unstructured.Unstructured) error {
	if strings.ToLower(u.GetKind()) != "pod" {
		return nil
	}

	p, err := unstructuredToPod(u)
	if err != nil {
		return err
	}

	ne.Pod = map[string]interface{}{}
	ne.Pod["start_time"] = p.Status.StartTime
	ne.Pod["ip"] = p.Status.PodIP
	ne.Pod["host_ip"] = p.Status.HostIP
	return nil
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
