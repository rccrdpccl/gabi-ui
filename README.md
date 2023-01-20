# Gabi UI

This projects enables [pgadmin](https://www.pgadmin.org/) to speak with [Gabi](https://github.com/app-sre/gabi).

## How to run it

### Spin up the applications

We need to export into `GABI_DOMAIN` environmental variable the target gabi domain (i.e. `https://gabi-domain.example.com`). Then we just run `docker compose up -d` and we will spin up `pgadmin` and a proxy that will talk to gabi.

```bash
export GABI_DOMAIN=<my target domain>

docker compose up -d
```

### Log into pgadmin
We can then hit `localhost:5050` from our favourite browser, and the login screen from pgadmin should show up.

We will use the declared usernme and password (see `docker-compose.yaml`) to log in.

### Connect to gabi

Once logged in we will already have a default server pointing to gabi proxy.
As a username, a mock username is defined. As password, we will be prompted to insert one once we try to connect to the declared gabi server. We should enter our Openshift token.

## Known issues

This software is not complete, and there are several known issues.

* Version of pgadmin doesn't official support Postgresql versions < 13
* Some features are not working as expected and might return errors
