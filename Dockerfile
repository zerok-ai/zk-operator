# Build the manager binary
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY opclients/ opclients/
COPY zk_yaml/ zk_yaml/ 

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu
RUN apt-get update 
RUN apt-get install -y git
ADD https://storage.googleapis.com/kubernetes-release/release/v1.9.4/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/zk_yaml/ zk_yaml/
USER 65532:65532
ENTRYPOINT ["/manager"]
