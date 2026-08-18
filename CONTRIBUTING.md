# Contributing

Requires **Go 1.22+**.

```bash
go test ./...
go vet ./...
```

On Linux/macOS also run `go test -race ./...` (needs cgo). GitHub Actions runs the race detector.

Keep public API changes backward compatible unless they belong in a new major version.

## Pull requests

- Include a test for parser or Codec 12 changes.
- Do not commit Teltonika protocol PDFs or other third-party documents. Link to the [Teltonika wiki](https://wiki.teltonika-gps.com/view/Codec) instead.
- Do not add sample packets that contain real production IMEI numbers unless they are already public fixtures in this repository.
