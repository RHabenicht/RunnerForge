# NSRCG II

Standalone Shadowrun Third Edition character generator for Windows 10/11, currently in development.

The Go application supports basic Priority character creation, local JSON saves, legacy .sr3 imports, and HTML export. Build Point creation, comprehensive rules validation, a finished installer, and Windows runtime acceptance testing remain pending.

The embedded catalog contains 6,449 records citing approved sources. Another 2,633 records are preserved in a separate review file. Source filtering does not establish that each record has been verified against its book.

## Development

Requires Go 1.23 or later and Python 3.

Run tests from the repository root.

```sh
python -m unittest discover -s . -p 'test_*.py'
cd go
go test ./...
go vet ./...
```

Build from the go directory on Windows.

```sh
go build -ldflags="-H windowsgui" -o NSRCG-II.exe ./cmd/nsrcgii
```

Recheck the catalog from the repository root.

```sh
python build_catalog.py --catalog go/srcapp/catalog.json
```

This reads both the embedded catalog and go/data/quarantine.json, preserving existing record IDs.

Four Python regression tests, Go tests, and go vet passed on November 6, 2026. Windows x86-64 cross-compilation succeeded. Native Windows GUI and PDF export testing remain outstanding.

DEVELOPMENT_STATUS.txt describes the latest checkpoint. Other historical reports describe earlier development stages.
