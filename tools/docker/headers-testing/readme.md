# Headers Testing

This docker setup allows for basic functional testing of the header striping logic

## QuickStart

Build and make the OmniSci WebServer and backend container
```
make build
```

Run basic functional testing to ensure the X-OmniSci-Username header is stripped

```
make test
```

To cleanup run

```
make clean
```
