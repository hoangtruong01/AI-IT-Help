# EOMP OpenAPI 3.0 hub

The canonical specification is [eomp-openapi-spec.yaml](eomp-openapi-spec.yaml).

Current coverage:

- OpenAPI version: 3.0.3
- Runtime operations: 101
- Documented operations: 101
- Coverage gate: `go run scripts/check_openapi_coverage.go`

The gate scans every `services/*/cmd/server/main.go`, checks that route patterns do not conflict under Go `http.ServeMux`, and compares the exact method/path set with this specification. Jenkins runs the same command.

The operation inventory is complete, but several write bodies and responses intentionally use generic JSON schemas. Domain-specific schemas and contract-conformance tests remain P1 work and are tracked in [the implementation status](../IMPLEMENTATION_STATUS.md).

To preview locally with Swagger UI:

```bash
docker run --rm -p 8088:8080 \
  -e SWAGGER_JSON=/spec/eomp-openapi-spec.yaml \
  -v "${PWD}/docs/openapi:/spec:ro" \
  swaggerapi/swagger-ui
```

Then open <http://localhost:8088>.
