FROM golang:1.26.2

ENV GOPROXY=off \
    GOSUMDB=off \
    GOTOOLCHAIN=local \
    CGO_ENABLED=0

WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /usr/local/bin/portcrane ./cmd/portcrane

EXPOSE 8080
CMD ["/usr/local/bin/portcrane", "-addr", "0.0.0.0:8080", "-data", "/tmp/portcrane"]
