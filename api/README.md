# API Specifications

The `/api` folder contains all specifications for the APIs exposed by `bbycrgo`.

## REST API

`bbycr.yaml` is an `OpenAPI 3.0` specification, generated using [Stoplight](https://stoplight.io/) and converted to code using [an OpenAPI Code Generator](https://github.com/deepmap/oapi-codegen):

```shell
oapi-codegen -config api/server.oapi-codegen.yaml api/bbycr.yaml
oapi-codegen -config api/types.oapi-codegen.yaml api/bbycr.yaml
```

It is then implemented in the `rest` package.
