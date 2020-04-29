package main

import (
	"log"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	eventCacheTable = "events"
)

type Handler struct {
	client *kubernetesClient
	ch     chan<- interface{}
	db     Cachier
}

func (h *Handler) OnAdd(obj interface{}) {
	var err error
	switch obj.(type) {
	case *v1.Event:
		event := obj.(*v1.Event)
		err = h.onEvent(event)
	case *v1.Service:
		err = h.onService(obj.(*v1.Service))
	}

	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) OnUpdate(oldObj, newObj interface{}) {
	var err error
	switch newObj.(type) {
	case *v1.Event:
		event := newObj.(*v1.Event)
		err = h.onEvent(event)
	case *v1.Service:
		err = h.onService(newObj.(*v1.Service))
	}

	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) OnDelete(obj interface{}) {
	var err error
	switch obj.(type) {
	case *v1.Event:
		event := obj.(*v1.Event)
		err = h.onEvent(event)
	case *v1.Service:
		err = h.onService(obj.(*v1.Service))
	}

	if err != nil {
		log.Println(err)
	}
}

func findServicePods(s *v1.Service) ([]interface{}, error) {
	return nil, nil
}

func (h *Handler) onService(s *v1.Service) error {
	// Save service to database
	// Maybe use s.SelfLink since UID is literally not exposed elsewhere for
	// a service other than the service itself being aware of it.
	// And a change in the service will change the UID afterall.
	if err := h.db.Set("service", string(s.UID), s); err != nil {
		return err
	}

	// Find all PODS for this service so that a rerverse lookup is possible.
	pods, err := findServicePods(s)
	if err != nil {
		return err
	}

	// Save service -> pod
	if err := h.db.Set("service-pods", string(s.UID), pods); err != nil {
		return err
	}

	// Also save pod -> service denormalized for reverse Index lookup
	for _, p := range pods {
		// A pod may be behind multiple services.
		// Get the existing array. append the new serviceID and set again
		// Calls for race condition probably. So will need some mutex here.
		if err := h.db.Set("pod-service", p.(string), s); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) onEvent(e *v1.Event) error {
	r, err := h.db.Get(eventCacheTable, string(e.UID))
	if err != nil {
		return err
	}

	// Event has been processed already.
	if r.Exists() {
		return nil
	}

	obj, err := h.client.getObject(h.db, &e.InvolvedObject)
	if err != nil {
		log.Println(err)
	}

	addr, err := h.client.getNodeAddress(h.db, e.Source.Host)
	if err != nil {
		log.Println(err)
	}

	h.ch <- makeL9Event(e, obj, addr)
	return nil
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
