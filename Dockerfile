FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o tailscale-ingress-operator main.go

# FROM gcr.io/distroless/static:nonroot
FROM golang:1.24
COPY --from=builder /app/tailscale-ingress-operator /
ENTRYPOINT ["/tailscale-ingress-operator"]