# Build the operator binary
FROM golang:1.18-alpine
WORKDIR /
COPY bin/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
