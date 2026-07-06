# SAML Testing Environment

An easy testing environment for SAML and the new webserver

## Quick Start

To start everything run

```
./build-saml.sh
```

Navigate to http://localhost:6273 and enter the license

### Setting up Keycloak SAML with Omnisci

1. Login to Keycloak at http://localhost:8081

1. On The Master Dropdown box click "Add Realm"

1. Select Import and import the "example-realm.json" file from this repo"

1. Press Create

Once thats done run the following script

```
./setup-saml.sh
```

This will configure OmniSci to use Keycloak Saml

### Setting up users

You'll need to add a SAML user.

1. Back in Keycloak Navigate to users on the bottom left

1. Click add user and enter the username only. I usually use (omnisci-saml)

1. Click Save

1. Navigate to the User > Credentials

1. Enter New password and hit reset password

1. Attempt to login to Omnisci so the user is created in the database

1. In omnisql ```ALTER USER omnisci-saml (is_super = 'true', default_db='omnisci');```

1. Log back in.

## Other Things

If you make changes to the webserver and want to build and test use:

```
./sync-webserver.sh
```