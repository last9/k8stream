# Setup guide to build and develop locally.

We advise developing against a [minikube](https://kubernetes.io/docs/tutorials/hello-minikube/)
Once you have a sample application going with Minikube. Proceed to the next stesp.

## Build

```bash
go build .
```
This should output a ./k8stream binary in the TLD of the repository.

## Run

```bash
./k8stream --config=config.json
```

## Configuration

Typical configuration looks like:

**NOTE** Strip the comments before consuming this JSON

```javascript
{
  "config": {
    "uid": "719395d7-4e91-4817-a6ec-9a8ded29bebc", // UID of this deployment
    "heartbeat_hook": "https://heartbeat.last9.io", // Heatbeat hook
    "heartbeat_interval": 60,     // Send a heartbeat signal.
    "batch_interval": 60,         // Flush every n seconds
    "batch_size": 10000,          // Flush every n events
    "sink": "memory"               // Choices "s3", "file", "memory"
  },
  "namespaces": ["default"],      // Skip this key if all namespaces should be captured. By default, kube-system, kubernetes, kubernetes-dashboard are always skipped

  // If the sink is "s3"
  "prefix": "local/test-upload",  // Prefix of S3 Upload
  "aws_region": "ap-south-1",     // Region of S3 bucket
  "aws_bucket": "last9-trials",   // S3 Bucket to Upload to
  "aws_profile": "last9data",     // Profile, in case using creds file

  // If the sink is "file"
  "file_sink_dir": "./logs",       // If the sink is "file"

  "kubeconfig": ""                // Location to kubeconfig file
}
```
