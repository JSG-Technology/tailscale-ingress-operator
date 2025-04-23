build:
  go build -o ingress-operator main.go

docker-build:
  docker build -t jgaudette/tailscale-ingress-operator:1.0.0 .

docker-push: docker-build
  docker push jgaudette/tailscale-ingress-operator:1.0.0

docker-run: docker-build
  docker run -it jgaudette/tailscale-ingress-operator:1.0.0

deploy: docker-push
  kubectl delete -f infra/deployment.yaml
  sleep 10
  kubectl apply -f infra/deployment.yaml
