# K8stream [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Build Status](https://travis-ci.com/last9/k8stream.svg?branch=master)](https://travis-ci.com/last9/k8stream)

![Logo|512x397, 50%](images/photo_2020-04-17_20-09-55.jpg)

# Background

Kubernetes events are an excellent source of information to monitor and debug the state of your cluster. Kubernetes API server emits events whenever there is a change in some resource it manages. These events are typically stored in etcd for some time and can be observed when you run kubectl get events or kubectl describe. 

Typical metadata in every event includes entity kind (pod, deployment etc), state (warning, normal, error), reason and message. 

Etcd is a fast key value store to retrieve these events but not for running analytics on top of it to figure out root causes of outages. This is where k8stream comes in as a pipeline to ingest events. 

# Overview

K8stream is a tool you can use to ingest Kubernetes events and send them to a specified sink in batches.

## Principles

- Goal is to enable storing events for post-hoc analysis
- The overhead to the cluster should be minimal
- All queries should be cached
- This processor does not handle deduplication and out of order events
- Events stored in the sink should be batched

## Non Goals

- This does not provide a UI or a queryable interface
- The storage is provided by the sink

# Setup

### Build

    make build

This should output a ./k8stream binary in the TLD of the repository.

### Run

    ./k8stream --config=config.json

### Configuration

Typical configuration looks like:

    {
        "batch_size": 5, # Flush every n events
        "batch_interval": 5, # Flush every n seconds
        "uid": "719395d7-4e91-4817-a6ec-9a8ded29bebc", # Unique Identifier to identify this stream in Sinks
        "file_sink_dir": "/tmp/l9k8stream/", # If the "sink" is configured to be a file
        "prefix": "local/test-upload", # Prefix for S3 Upload
        "aws_region": "ap-south-1", # Region of S3 Bucket
        "aws_bucket": "last9-trials", # S3 Bucket to Upload to
        "aws_profile": "last9data", # AWS Profile reads from ~/.aws/credentials
        "sink": "file", # Should use S3 of File Sink
        "kubeconfig": "./kubeconfig" # Location to kubeconfig file, leave empty when deploying to K8s
    }

### Deploy

Sample-config file at 

    apiVersion: v1
    kind: Namespace
    metadata:
      name: last9
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      namespace: last9
      name: k8stream
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: k8stream
    # These rules will be added to the "monitoring" role.
    rules:
    - apiGroups: ["*"]
      resources: ["services", "endpoints", "pods", "nodes", "events", "deployments", "replicasets"]
      verbs: ["get", "list", "watch"]
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: k8stream
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: k8stream
    subjects:
      - kind: ServiceAccount
        namespace: last9
        name: k8stream
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: k8stream-config
      namespace: last9
    data:
      config.json: |
        {
          "batch_size": 5,
          "batch_interval": 5,
          "prefix": "local/test-upload",
          "uid": "719395d7-4e91-4817-a6ec-9a8ded29bebc",
          "file_sink_dir": "/tmp/l9k8stream/",
          "aws_region": "ap-south-1",
          "aws_bucket": "last9-trials",
          "aws_profile": "last9data",
          "aws_access_key": "1",
          "aws_secret_access_key": "2",
          "sink": "file",
          "kubeconfig": ""
        }
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: k8stream
      namespace: last9
    spec:
      replicas: 1
      template:
        metadata:
          labels:
            app: k8stream
            version: v1
        spec:
          serviceAccountName: k8stream
          restartPolicy: Always
          containers:
            - name: k8stream
              imagePullPolicy: Never
              image: last9inc/k8stream:a14a00f
              command: ["/app/agent"]
              args: ["--config", "/data/config.json"]
              volumeMounts:
                - mountPath: /data
                  name: cfg
          volumes:
            - name: cfg
              configMap:
                name: k8stream-config
      selector:
        matchLabels:
          app: k8stream
          version: v1


In case on in-cluster deployment omit the "kubeconfig" parameter in JSON. Setting this as empty the code falls back to in-cluster authorization.

# Detailed Design
![](images/k8stream.jpg)


## Handling of events

- Doesn't perform an event-by-event upload but instead uploads in batches
- K8s stream handles some basic de-duplication of events by checking the event cache if this entry has been processed.  However, if the cache gets flushed or the k8stream binary gets restarted, it will start processing duplicate events.
- The events are enriched with more metadata such as labels and annotations corresponding to the entity, node IP address etc.
- Uses a highly concurrent Cache to avoid re-lookup.

## Writing to sink

- Uses asynchronous batching to write to Sink (only S3, File outout are supported for now)
- Events are marshalled using protobuf.
- Data written to sink is gzipped.
- Avoids any local/intermediate files.
- Resync allows to catch up with the Event stream if its lost momentarily.

# Limitations

- Because events from K8s can arrive out of order, though we try our best to de-deduplicate and order them, it cannot be guaranteed. It's advised to handle deduplication and ordering at consumer end.
- K8stream does not handle the case of duplicate events after  a restart. This is because the only deduplication that happens currently is by reading the local cache which gets flushed on a restart. This needs to be handled by the consumer of the stream.
- This currently only supports writing to S3 and file output.

# Future Work

- Support for writing to more output streams
- Adding more metadata like service name and pod IP address for the event
