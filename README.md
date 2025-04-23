# Tailscale Ingress Operator

A Kubernetes operator that automatically creates Tailscale Ingress resources for services annotated with `jsgtechnology.com/tailscale-autoingress`.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

This operator watches for Kubernetes services and automatically creates Tailscale Ingress resources when a service is annotated with `jsgtechnology.com/tailscale-autoingress`. It handles the creation, updates, and deletion of Ingress resources based on service lifecycle events.

## Features

- Automatically creates Tailscale Ingress resources for annotated services
- Handles service updates and deletions
- Supports TLS configuration
- Uses the first service port for the Ingress backend
- Creates Ingress resources with the `tailscale` IngressClass

## Prerequisites

- Kubernetes cluster (tested with v1.24+)
- Tailscale Ingress Controller installed in the cluster
- Go 1.24+ for building from source

## Installation

### Using Helm

```bash
helm repo add tailscale-ingress-operator https://your-helm-repo
helm install tailscale-ingress-operator tailscale-ingress-operator/tailscale-ingress-operator
```

### Manual Installation

1. Build the operator:
```bash
go build -o tailscale-ingress-operator main.go
```

2. Deploy to your cluster:
```bash
kubectl apply -f infra/deployment.yaml
```

## Usage

To enable automatic Ingress creation for a service, add the following annotation:

```yaml
metadata:
  annotations:
    jsgtechnology.com/tailscale-autoingress: "true"
```

Example service manifest:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
  annotations:
    jsgtechnology.com/tailscale-autoingress: "true"
spec:
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app: my-app
```

The operator will automatically create an Ingress resource named `my-service-ingress` with the following characteristics:
- Uses the `tailscale` IngressClass
- Forwards traffic to the first port defined in the service
- Configures TLS with the service name as the host

## Development

### Building

```bash
go build -o tailscale-ingress-operator main.go
```

### Running Locally

```bash
go run main.go
```

### Building Docker Image

```bash
docker build -t tailscale-ingress-operator:latest .
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a history of changes to this project. 