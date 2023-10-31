## Description

Example of a simple CRUD microservice written in Go.



## Assumptions

- JWT-token is provided with each request
- Users can't delete themselves
- Users can have permissions to create/delete/update other users
- The repository uses soft deletion of user records



## Install tools for linting, code generation, etc.

```bash
make install-tools
```



## Build

```bash
make build
```

Set the following environmental variables before running the service:

```bash
USERS_LOG_LEVEL (possible values: debug, info, warn, error, panic, fatal; default is "info")
USERS_REPOSITORY_URI
USERS_JWT_SECRET
USERS_JWT_ISSUER
USERS_HTTP_BIND_ADDRESS (default is ":8080")
```



## Running locally

```bash
docker-compose up
```

HTTP port `8088` is exposed by the container.

Go to `localhost:8088/docs` to see the OpenAPI specification for the available endpoints.

For the test purposes, the following `JWT`-token can be used: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyX2F1dGhlbnRpY2F0aW9uIiwidXNlcl9pZCI6MTIzNDU2Nzg5LCJjYW5fY3JlYXRlX3VzZXJzIjp0cnVlLCJjYW5fZGVsZXRlX3VzZXJzIjp0cnVlLCJjYW5fdXBkYXRlX3VzZXJzIjp0cnVlLCJjYW5fdmlld191c2VycyI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE3NjE4OTg2MjEsImlzcyI6ImxvY2FsaG9zdCJ9.HVQMV6ENzpU8SIBPi_fsBn_d5FdyW1ej-a0_0qkAYqA` (user with full permissions to create, view, update, delete users).
