# Weberver using Fips140-2 BoringCypto

The included scripts/dockerfile will build a fips 140-2 compliant(not certified) binary with BoringCrypto.

1. FIPS 140-2 Information
    - Boringcrypto license: https://boringssl.googlesource.com/boringssl/+/refs/tags/fips-20190808/LICENSE
    - Boringcrypto repo: https://boringssl.googlesource.com/boringssl/+/refs/tags/fips-20190808
    - FIPS 140-2 certificate on NIST’s webpage: https://csrc.nist.gov/projects/cryptographic-module-validation-program/Certificate/2964
    - Full FIPS 140-2 certificate: https://csrc.nist.gov/CSRC/media/projects/cryptographic-module-validation-program/documents/security-policies/140sp2964.pdf
1. Golang with boringcrypto
    - Repo: https://go.googlesource.com/go/+/dev.boringcrypto

## Building

Run make then check the build-artifacts dir

```
make
```

Cleanup can be done with

```
make clean
```

Tests can be run with

```
make test
```