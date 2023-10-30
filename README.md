## Description

Example of a simple CRUD microservice written in Go.



## Assumptions

- JWT-token is provided with each request
- Users can't delete themselves
- Users can have permissions to create/delete/update other users



## Install tools for linting, code generation, etc.

```bash
make install-tools
```



## Build

```bash
make build
```



## Run

```bash
./users.bin -c config.yaml
```

