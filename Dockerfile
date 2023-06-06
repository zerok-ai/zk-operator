# Build the operator binary
FROM alpine:latest
WORKDIR /
COPY bin/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
