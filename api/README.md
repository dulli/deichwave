# API Specifications

The `/api` folder contains all specifications for the APIs exposed by `bbycrgo`.

## REST API

`bbycr.yaml` is an `OpenAPI 3.0` specification, generated using [Stoplight](https://stoplight.io/) and converted to code using [an OpenAPI Code Generator](https://github.com/deepmap/oapi-codegen):

```shell
oapi-codegen -generate chi-server -o pkg/rest/openapi_server.gen.go -package rest api/bbycr.yaml
oapi-codegen -generate client -o pkg/rest/openapi_client.gen.go -package rest api/bbycr.yaml
oapi-codegen -generate types -o pkg/rest/openapi_types.gen.go -package rest api/bbycr.yaml
```

It is then implemented in the `rest` package.
