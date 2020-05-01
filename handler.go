package main

import (
	fmt "fmt"
	"log"

	v1 "k8s.io/api/core/v1"
)

const (
	serviceTable     = "service"
	eventCacheTable  = "events"
	servicePodsTable = "service-pods"
	podServicesTable = "pod-service"
	serviceAppsTable = "service-apps"
	appServicesTable = "apps-service"
)

type Handler struct {
	client *kubernetesClient
	ch     chan<- interface{}
	db     Cachier
	conf   *L9K8streamConfig
}

func (h *Handler) OnAdd(obj interface{}) {
	var err error
	switch obj.(type) {
	case *v1.Event:
		event := obj.(*v1.Event)
		err = h.onEvent(event)
	case *v1.Service:
		err = h.onService(obj.(*v1.Service), "addedService")
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
		err = h.onService(newObj.(*v1.Service), "updatedService")
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
		err = h.onService(obj.(*v1.Service), "deletedService")
	}

	if err != nil {
		log.Println(err)
	}
}

func contains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

var skipNamespaces = []string{"kube-system", "kubernetes", "kubernetes-dashboard"}

func (h *Handler) onService(s *v1.Service, eventType string) error {
	// Do not watch the default kubernetes services
	switch {
	case contains(s.GetNamespace(), skipNamespaces):
		return nil
	case len(h.conf.Namespaces) > 0 && !contains(s.GetNamespace(), h.conf.Namespaces):
		return nil
	default:
		if s.GetName() == "kubernetes" {
			return nil
		}
	}

	suid := string(s.GetUID())
	eventId := fmt.Sprintf("%s-%s", suid, s.GetResourceVersion())

	r, err := h.db.Get(eventCacheTable, eventId)
	if err != nil {
		return err
	}

	// Service has been processed already.
	if r.Exists() {
		var existingService L9Event
		if err := r.Unmarshal(&existingService); err != nil {
			return err
		}

		// Ignore if event is already processed.
		if existingService.ReferenceVersion >= s.GetResourceVersion() {
			log.Println("Service", suid, "already processed")
			return nil
		}
	}

	event, err := makeL9ServiceEvent(h.db, h.client, eventId, s, eventType)
	if err != nil {
		return err
	}

	h.ch <- event
	return nil
}

func (h *Handler) isEligible(obj *v1.Event) bool {
	if contains(obj.Namespace, skipNamespaces) {
		return false
	}
	return (len(h.conf.Namespaces) == 0 || contains(obj.Namespace, h.conf.Namespaces)) && (len(h.conf.Events) == 0 || contains(obj.Reason, h.conf.Events))
}

func (h *Handler) onEvent(e *v1.Event) error {
	if !h.isEligible(e) {
		return nil
	}

	r, err := h.db.Get(eventCacheTable, string(e.UID))
	if err != nil {
		return err
	}

	// Event has been processed already.
	if r.Exists() {
		return nil
	}

	event, err := makeL9Event(h.db, h.client, e)
	if err != nil {
		return err
	}

	h.ch <- event
	return nil
}
