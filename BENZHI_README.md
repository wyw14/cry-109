# PortCrane

PortCrane is a local control service for quay-crane lift, spreader, trolley,
heave compensation, rail motion, braking and storm anchoring operations.

Run the service with Go 1.26.2:

```sh
go run -mod=vendor ./cmd/portcrane -addr 127.0.0.1:8080 -data ./data
```

The operator pages are available at `/lifts`, `/spreader`, `/motion` and
`/safety`. The health endpoint is `/healthz`; JSON control endpoints are under
`/api`.

The container build is intentionally offline and uses the committed vendor
tree:

```sh
docker build -f benzhi.Dockerfile -t portcrane:local .
```
