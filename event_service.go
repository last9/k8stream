package main

import (
	"encoding/json"
	"time"

	v1 "k8s.io/api/core/v1"
)

func cacheServicePods(c *kubernetesClient, db Cachier, s *v1.Service) ([]v1.Pod, error) {
	suid := string(s.GetUID())

	// Find all PODS for this service so that a rerverse lookup is possible.
	pods, err := c.getPods(db, s)
	if err != nil {
		return pods, err
	}

	// Save service -> pods
	if err := db.Set(servicePodsTable, suid, pods); err != nil {
		return pods, err
	}

	// Also save pod -> service denormalized for reverse Index lookup
	for _, p := range pods {
		// A pod may be behind multiple services.
		// Get the existing array. append the new serviceID and set again
		// Calls for race condition probably. So will need some mutex here.
		if err := db.Set(
			makeKey(podServicesTable, string(p.GetUID())), suid, true,
		); err != nil {
			return pods, err
		}
	}

	return pods, nil
}

func cacheServiceApps(c *kubernetesClient, db Cachier, s *v1.Service) error {
	suid := string(s.GetUID())

	// Find all Replication Controllers for this service
	// so that a rerverse lookup is possible.
	apps, err := c.getApps(db, s)
	if err != nil {
		return err
	}

	// Save service -> ReplicationControllers
	if err := db.Set(serviceAppsTable, suid, apps); err != nil {
		return err
	}

	// Also save rc -> service denormalized for reverse Index lookup
	for _, r := range apps {
		// A pod may be behind multiple services.
		// Get the existing array. append the new serviceID and set again
		// Calls for race condition probably. So will need some mutex here.
		if err := db.Set(
			makeKey(appServicesTable, string(r.GetUID())), suid, true,
		); err != nil {
			return err
		}
	}

	return nil
}

// eventID
func makeL9ServiceEvent(db Cachier, c *kubernetesClient, eventID string, s *v1.Service, eventType string) (*L9Event, error) {
	suid := string(s.GetUID())

	// Save service to database
	// Maybe use s.SelfLink since UID is literally not exposed elsewhere for
	// a service other than the service itself being aware of it.
	// And a change in the service will change the UID afterall.
	if err := db.Set(serviceTable, suid, s); err != nil {
		return nil, err
	}

	pods, err := cacheServicePods(c, db, s)
	if err != nil {
		return nil, err
	}

	if err := cacheServiceApps(c, db, s); err != nil {
		return nil, err
	}

	podMap := map[string]interface{}{}
	for _, p := range pods {
		b, err := json.Marshal(miniPodInfo(p))
		if err != nil {
			podMap[p.GetName()] = err.Error()
		} else {
			podMap[p.GetName()] = string(b)
		}
	}

	return &L9Event{
		ID:                 eventID,
		Timestamp:          time.Now().Unix(),
		Component:          s.GetName(),
		Host:               "",
		Message:            eventType,
		Namespace:          s.GetNamespace(),
		Reason:             eventType,
		ReferenceUID:       "",
		ReferenceNamespace: "",
		ReferenceName:      "",
		ReferenceKind:      "",
		ReferenceVersion:   s.GetResourceVersion(),
		ObjectUid:          string(s.GetUID()),
		Labels:             s.GetLabels(),
		Annotations:        s.GetAnnotations(),
		Address:            nil,
		Pod:                podMap,
		Services:           []string{s.GetName()},
	}, nil
}
