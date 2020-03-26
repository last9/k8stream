# k8stream [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Build Status](https://travis-ci.com/last9/k8stream.svg?branch=master)](https://travis-ci.com/last9/k8stream)
Process & Send kubernetes events to a Sink, in batches.

- Doesn't perform an event-by-event upload.
- Uses asynchronous batching to write to Sink (only S3, File outout are supported for now)
- Events are marshalled using protobuf.
- Data written to Sink is gzipped.
- Handles de-duplication to some degreee. Not applicable after a restart.
- Enriches Event with Object and Host details and Caches them.
- Uses a highly concurrent Cache to avoid re-lookup.
- Avoids any local/intermediate files.
- Resync allows to catch up with the Event stream if its lost momentiarilly.


*Because events from K8s can arrive out of order, though we try our best to de-deuplicate and order them, it cannot be guaranteed. It's advised to handle deduplication and ordering at consumer end*

## Build

Follow the Makefile `make build`
It should output a ./k8stream binary in the TLD of the repository.

## Run

`./k8stream --config=config.json`

## Configuration

```python
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
```

## Deploy

To deploy refer to the sample-config file at https://github.com/last9/k8stream/blob/master/deploy/k8stream.yaml
In case on in-cluster deployment omit the "kubeconfig" parameter in JSON. Setting this as empty the code falls back to in-cluster authorization.
