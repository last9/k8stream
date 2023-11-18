<a href="https://last9.io"><img src="https://last9.github.io/assets/last9-github-badge.svg" align="right" /></a>

# K8stream [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Build Status](https://travis-ci.com/last9/k8stream.svg?branch=master)](https://travis-ci.com/last9/k8stream)

![Logo|512x397, 50%](images/photo_2020-04-17_20-09-55.jpg)

# Background

Kubernetes events are an excellent source of information to monitor and debug
the state of your Services. Kubernetes API server emits events whenever there is
a change in some resource it manages. These events are typically stored in etcd
for some time and can be observed when you run kubectl get events or kubectl
describe.

Typical metadata in every event includes entity kind (pod, deployment etc),
state (warning, normal, error), reason and message.

There are no tools for running analytics on top of it to figure out root causes
of outages. This is where k8stream comes in as a pipeline to ingest events.

# Overview

K8stream is a tool you can use to:

- Ingest Kubernetes events to a Sink for Offline analytics
- **Find correlation with services under Impact**

## Principles

- Goal is to enable storing events for post-hoc analysis
- Pods are cattle, Services are Pets. Any change in cluster should find its
  association with the Service under Impact.
- The overhead to the cluster should be minimal
- All queries should be cached
- Events stored in the sink should be batched
- Configuration mode for Duplicates-Events vs No-Event-Loss

## Non Goals

- This does not provide a UI or a queryable interface
- The storage is provided by the sink

# Deploy

There are sample deployment files available at [K8s YAMLs](deploy/)

In case on in-cluster deployment omit the "kubeconfig" parameter in JSON.
Setting this as empty the code falls back to in-cluster authorization.

# Detailed Design

![](images/k8stream.jpg)

## Handling of events

- Doesn't perform an event-by-event upload but instead uploads in batches
- K8stream handles some basic de-duplication of events by checking the event
  cache if this entry has been processed. However, if the cache gets flushed or
  the k8stream binary gets restarted, it will start processing duplicate events.
- The events are enriched with more metadata such as labels and annotations
  corresponding to the entity, node IP address etc.
- Uses a highly concurrent Cache to avoid re-lookup.

## Writing to sink

- Uses asynchronous batching to write to Sink (only S3, File outout are
  supported for now)
- Events are marshalled using protobuf.
- Data written to sink is gzipped.
- Avoids any local/intermediate files.
- Resync allows to catch up with the Event stream if its lost momentarily.

# Limitations

- Because events from K8s can arrive out of order, though we try our best to
  de-deduplicate and order them, it cannot be guaranteed. It's advised to handle
  deduplication and ordering at consumer end.
- K8stream does not handle the case of duplicate events after a restart. This is
  because the only deduplication that happens currently is by reading the local
  cache which gets flushed on a restart. This needs to be handled by the
  consumer of the stream.
- This currently only supports writing to S3 and file output.

# Future Work

- [x] Support for adding POD details to an event
- [x] Support for writing to more output streams
- [x] Adding more metadata like service name for the event

# Development and Contribution

Please follow the [Developer](Development.md) guide to setup and run k8stream in
a local environment.

---

# About Last9

This project is sponsored by [Last9](https://last9.io). Last9 builds reliability tools for SRE and DevOps.

<a href="https://last9.io"><img src="https://last9.github.io/assets/email-logo-green.png" alt="" loading="lazy" height="40px" /></a>
