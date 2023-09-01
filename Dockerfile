# Build the operator binary
FROM alpine:latest
WORKDIR /
COPY bin/manager .
ENTRYPOINT ["/manager"]
