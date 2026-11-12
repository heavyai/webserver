# Updating the Thrift Bindings

Install the same version of the Thrift compiler used by `go.mod`, then clone
the HeavyDB repository.

## Manually

For Thrift 0.25.0, verify the compiler and regenerate the bindings with:

```bash
thrift --version
cd "${HEAVY_DB_DIR}"
rm -rf gen-go
thrift -r \
  --gen go:package_prefix=github.com/heavyai/webserver/internal/db_thrift_client/ \
  heavy.thrift
for pkg in gen-go/*/; do
  rsync -a --delete "$pkg" \
    "${WEBSERVER_DIR}/internal/db_thrift_client/$(basename "$pkg")/"
done
```

The compiler version must match the Thrift library version in `go.mod`.
Regeneration preserves the hand-written
`internal/db_thrift_client/db_thrift_client.go` file.
