## Pod Exporter

A light-weight golang web service that scrapes the kubernetes api for pod metadata.

Additionally, it exposes prometheus metrics on the `/metrics` endpoint; to track count, of both recent and pods older than a week.

### Folder structure
```
.
├── Dockerfile
├── cmd/
│   ├── go.mod
│   ├── go.sum
|   ├── main.go
|   ├── metrics.go
│   └── pods.go
├── kind/
│   └── kind-cluster.yml
└── kubernetes/
    ├── ingress.yml
    ├── namespace.yml
    ├── deployment.yml
    ├── rbac.yml
    └── service.yml
```
### Endpoints
* `/metrics`: returns prometheus metrics
    ```text
    # HELP pods_recent_count_current The total number of pods older or younger than 7 days
    # TYPE pods_recent_count_current gauge
    pods_recent_count_current{valid="false"} 0
    pods_recent_count_current{valid="true"} 3
    ```
### Setup
Install golang dependencies using the `go mod` utility:
```
$ cd cmd
$ go mod tidy
```

```
$ go get github.com/prometheus/client_golang/prometheus/promauto
$ go get github.com/prometheus/client_golang/prometheus
$ go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Building
The `Dockerfile` packages the service into an alpine based docker image. This keeps the image size small and quick to deploy:

```
docker build -t eugenebad/pod-exporter:0.0.1c .
```
Push image:
```
docker push eugenebad/pod-exporter:0.0.1c
```

> Inorder to push to a different docker registry, rename the image accordingly:
> ```docker build -t <registry_name>/pod-exporter:0.0.1c .```

### Running on kubernetes
The service can be deployed on a kubernetes cluster using the manifest files in `/kubernetes`

Such a test environment can be bootstrapped locally using the [`kind`](https://kind.sigs.k8s.io/) utility.

##### 1. Installation (on Linux)
Download and install the kind binary
```
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.17.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

##### 2. Create kind cluster
From repo directory:
```
$ kind create cluster --config kind/kind-cluster.yml
```
> Note: This sets your current kubectl context to be `kind-kind`

##### 3. Install Nginx ingress controller
Inorder to route external requests to services running inside the cluster:
```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```
##### 4. Deploy the pod-exporter service
1. Create namespace:
```
$ kubectl create -f kubernetes/namespace.yml
```
Before deploying the application, note that the `kubernetes/ingress.yml` file has a **host** attribute which defines the domain name on which the service will be accessed. You can set this to your liking, so long as it resolves to the local address `127.0.0.1`. Which can be achieved by using the `/etc/hosts` file.

> Ideally, for non-local deployment, dns resolution should be handled by a dedicated DNS server.

2. Deploy application:
```
$ kubectl apply -f kubernetes/
```

This exposes the service on `localhost` at port `80`

```
$ curl localhost/metrics
```

##### 5. Logging
The service logs json output to stdout. Obtain the deployed pod using:
```
$ kubectl get po -n pod-exporter
```
and then tail the logs:
```
$ kubectl logs -f -n pod-exporter -p pod-exporter-7cd7bd6bd4-8s6gs -c pod-exporter --previous=false
{"level":"info","msg":"Application started successfully! Listening on port 9090...","time":"2023-05-15T17:51:43Z"}

{"level":"info","msg":"","pod":"money-5949c664bb-pt6fm","rule_evaluation":[{"name":"image_prefix","valid":false},{"name":"team_label_present","valid":false},{"name":"recent_start_time","valid":true}],"time":"2023-05-15T18:02:43Z"}
{"level":"info","msg":"","pod":"nomoney-7bcfdfbcb8-kz6tf","rule_evaluation":[{"name":"image_prefix","valid":false},{"name":"team_label_present","valid":true},{"name":"recent_start_time","valid":true}],"time":"2023-05-15T18:02:43Z"}
{"level":"info","msg":"","pod":"test-pod","rule_evaluation":[{"name":"image_prefix","valid":false},{"name":"team_label_present","valid":true},{"name":"recent_start_time","valid":true}],"time":"2023-05-15T18:02:43Z"}
```
##### 6. Cleanup

```
kind delete cluster
```