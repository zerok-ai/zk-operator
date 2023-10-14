# Build the operator binary
FROM alpine:latest
WORKDIR /
COPY bin/manager-amd64 .
COPY *bin/manager-arm64 .

COPY app-start.sh /app-start.sh
RUN chmod +x /app-start.sh

ENV ARGS "--health-probe-bind-address=:8081 --metrics-bind-address=127.0.0.1:8080 --leader-elect"

CMD ["./app-start.sh","-amd64","manager-amd64 $ARGS","-arm64","manager-arm64 $ARGS"]

