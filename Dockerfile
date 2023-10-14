# Build the operator binary
FROM alpine:latest
WORKDIR /
COPY bin/manager-amd64 .
COPY *bin/manager-arm64 .

COPY app-start.sh /app-start.sh
RUN chmod +x /app-start.sh


ENTRYPOINT ["./app-start.sh","-amd64","manager-amd64","-arm64","manager-arm64"]
