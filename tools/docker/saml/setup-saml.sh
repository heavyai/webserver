#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


tee ./omniscidata/omnisci.conf > /dev/null <<EOF
port = 6274
http-port = 6278
calcite-port = 6279
data = "/omnisci-storage/data"
null-div-by-zero = true
saml-metadata-file = "/omnisci-storage/omnisci-saml.xml"
saml-sp-target-url = "saml-testing"
allow-local-auth-fallback = 1
saml-sync-roles = 1

[web]
port = 6273
servers-json = "/omnisci-storage/servers.json"
EOF

tee ./omniscidata/omnisci-saml.xml > /dev/null <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="http://localhost:8081/auth/realms/saml-testing">
   <md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
      <md:KeyDescriptor use="signing">
      <dsig:KeyInfo xmlns:dsig="http://www.w3.org/2000/09/xmldsig#">
            <dsig:X509Data>
              <dsig:X509Certificate>MIICpzCCAY8CBgFvGwWG4zANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDDAxzYW1sLXRlc3RpbmcwHhcNMTkxMjE4MjE1NzI0WhcNMjkxMjE4MjE1OTA0WjAXMRUwEwYDVQQDDAxzYW1sLXRlc3RpbmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCy3LSW9RKkXg+Ve8lzkKu8k25NArUeJz6hdEf5ZxO1odQkvqWXRhKFCUR4Pd66/AMHiSoHNsnLhr48RAwJ+5bTqxHuaZmV4ZeP1Ytudb/jw+iSt07UZw+Uxy8NiUU74pHkOXYCgpV6zbL7mjJQHDGEG9NAdpMV3fO8s7WHqLD5Istuq8dOELvMMi9dBtXTv9ZPFN9JoX7tM/+ZIMKPfL0ChM68zV7lvNci1KWUTzxe2mOx6l7xnjuuf0BQzmAnl+66f2Yb4LrIAuBElX/U8GaUI+TR+goRCfzrqmiZsYdBAh0adEluh6Nvm8NGhyCe2oofJR1OUdHq4d7Lm4CMkbI1AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAIsdmgmwbBZje54lgvzyJ0ZgTmsP0dDguHNx1RrHhvXRzWjY+bSy69tVffKvC1KZ32/7LfQG+cY56DOi1zjT2tlrkZoIIzAxxxYZhzN2BzgKtEIy8W/3G+PAKoUx99c4i/GMlEDY8sveHV1VJg3FKDwiTBZ6BxhXNGkPNfQFmh7GYgaCkHKTkswQwHG4bYXBP2pTW6SPQBF3F4EHRpvt+6IKdSTuQsNEn5ypV2mavHSG0mKGX2z7JQZIHiOd2Ck9zS88KsIpSUzqkUPQh8kCK+yoCZrvFL1BTAaZcY5HxDqUdYqUqsC/WhIE04JgtEJLd8xSrviexvyVVmV7fgu6gIE=</dsig:X509Certificate>
            </dsig:X509Data>
          </dsig:KeyInfo>
      </md:KeyDescriptor>
      <md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>
      <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="http://localhost:8081/auth/realms/saml-testing/protocol/saml" />
      <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="http://localhost:8081/auth/realms/saml-testing/protocol/saml" />
   </md:IDPSSODescriptor>
</md:EntityDescriptor>
EOF

tee ./omniscidata/servers.json > /dev/null <<EOF
[
  {
    "database": "omnisci",
    "host": "localhost",
    "master": true,
    "password": "HyperInteractive",
    "port": 6273,
    "username": "admin",
    "SAMLurl":"http://localhost:8081/auth/realms/saml-testing/protocol/saml/clients/saml-testing"
  }
]
EOF

docker-compose restart omnisci-server
