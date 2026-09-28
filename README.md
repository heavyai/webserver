# heavyai-webserver
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/heavyai/webserver/blob/main/LICENSE.txt)
[![Security](https://img.shields.io/badge/Security-Report%20a%20Vulnerability-red.svg)](https://github.com/heavyai/webserver/blob/main/SECURITY.md)
[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-blue?logo=github)](https://github.com/orgs/heavyai/discussions)



Application web server for HeavyDB client-side application integration

## Installation
- Get Golang development environment up and running. >= v1.13
- Clone project into folder under $GOPATH
- Run `./scripts/build.sh` in project root dir
- Run `./build/heavy_web_server -b [Core root URL]`

##### Example run command:
`./build/heavy_web_server -bhttp://forge.mapd.com:6278`

Which ever instance of `heavydb` you're pointing to w/ `-b`, make sure you're
pointing at the port Thrift is listening to (`6278` on forge), _not_ the OWS
port (`6273`)

One can provide a path to a `.pem` file containing the RSA key used to encrypt
JWTs, if a file is not provided, a random key will be generated:
`./build/heavy_web_server -bhttp://forge.mapd.com:9092 --jwt-key-file [path to RSA pem key file]`

Server should serve on port `6273`

Run `curl -X GET -i http://localhost:6273/` to see the ✨🔮

*NOTE: you may need to pass the `--backend-url ...` and `--frontend <path to
compiled Immerse>` options to the server to see it function properly. If you
get a hard crash trying to hit the server, ensure `--frontend` is set properly
to point to Immerse's `dist` directory, and you have compiled Immerse.*

## Configuration

Most settings can be provided via CLI flag, environment variable (`HEAVY_<SECTION>_<FLAG-NAME>`, e.g. `HEAVY_WEB_MAPBOX-TOKEN` — note hyphens in flag names are preserved, not converted to underscores), or a TOML config file passed with `--config`/`-c` (see `tools/docker/local/heavy.conf` for an example). Web-server-specific settings live under a `[web]` table in that file.

### Map provider keys

To enable map charts in Immerse without rebuilding the frontend, set your own Mapbox and/or Google Maps API keys in `heavy.conf`:

```toml
[web]
mapbox-token = "pk.<your Mapbox access token>"
google-api-key = "<your Google Maps API key>"
```

These are injected into Immerse at request time — no rebuild or redeploy of the frontend bundle required. As with any API key, restrict it to your domain in the Mapbox/Google dashboard; the key is still visible to anyone using the running app.

## Tools
See `/tools` for various development and testing tools with separate dedicated READMEs.

## Third-Party Licenses

A full list of third-party Go modules and their licenses is maintained in [`third_party_licenses/THIRD_PARTY_LICENSES.md`](third_party_licenses/THIRD_PARTY_LICENSES.md). To regenerate it after dependency changes, run:

```sh
./scripts/generate-third-party-licenses.sh
```

This requires `go mod download` to have been run so modules are present in the local cache.

## Security
> [!WARNING]
> **Do not report security vulnerabilities through public GitHub issues!**

NVIDIA takes security seriously. If you discover a vulnerability in webserver, **DO NOT open a public issue**. Use one of the private reporting channels described in [SECURITY.md](https://github.com/heavyai/webserver/blob/main/SECURITY.md).

## Support
Join the [HeavyAI GitHub Discussions](https://github.com/orgs/heavyai/discussions) to ask questions, share feedback, and report issues. HeavyAI maintainers review issues, discussions, and pull requests on a best effort basis without guaranteed response timelines.
  
## License
Apache 2.0. See [LICENSE](https://github.com/heavyai/webserver/blob/main/LICENSE.txt).

