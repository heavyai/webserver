# Local Jupyter Testing Environment

A locally deployed testing environment for using Immerse with Jupyter, backed by an included OmniSciDB instance in CPU mode.

Intended for local web server development.

## Requirements

Requires Docker to be installed locally.

## Start

1. Add `enableJupyter` and the default OmniSciDB credentials to `servers.json` with `url` set to the local web server. Example:

```
[
  {
    "username": "admin",
    "password": "HyperInteractive",
    "database": "omnisci",
    "enableJupyter": true,
    "url": "http://localhost:6273",
    "master": true
  }
]
```

2. Run `make deploy` in this directory to start the local Jupyter+OmniSciDB service.

3. Build and run a local web server with the backend URL set to `http://localhost:6278` and the
   Jupyter URL set to to `http://localhost:8000`. Example:

```
$ /build/heavy_web_server --frontend ../immerse/dist --servers-json ../immerse/src/servers.local.json --verbose=true  --development=true --backend-url http://localhost:6278 --jupyter-url http://localhost:8000
```

## Stop

Stop the local Jupyter service by running `make clean` in this directory.
