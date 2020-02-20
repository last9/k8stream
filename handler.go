package main

import (
	"log"

	"github.com/dgraph-io/ristretto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Handler struct {
	client *kubernetesClient
	ch     chan<- *L9Event
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

func makeL9Event(
	e *v1.Event, u *unstructured.Unstructured, address []string,
) *L9Event {
	ne := &L9Event{}
	ne.ID = string(e.UID)
	ne.Timestamp = e.CreationTimestamp.Time.Unix()
	ne.Component = e.Source.Component
	ne.Host = e.Source.Host
	ne.Message = e.Message
	ne.Namespace = e.Namespace
	ne.Reason = e.Reason
	ne.ReferenceUID = string(e.InvolvedObject.UID)
	ne.ReferenceName = e.InvolvedObject.Name
	ne.ReferenceVersion = e.InvolvedObject.APIVersion
	if u != nil {
		ne.ReferenceNamespace = u.GetNamespace()
		ne.ReferenceKind = u.GetKind()
		ne.ObjectUid = string(u.GetUID())
		ne.Labels = u.GetLabels()
		ne.Annotations = u.GetAnnotations()
	}
	ne.Address = address
	return ne
}
