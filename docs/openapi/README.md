# EOMP OpenAPI 3.0 hub

The canonical specification is [eomp-openapi-spec.yaml](eomp-openapi-spec.yaml).

Current coverage:

- OpenAPI version: 3.0.3
- Runtime operations: 107
- Documented operations: 107
- Coverage gate: `go run scripts/check_openapi_coverage.go`

The gate scans every `services/*/cmd/server/main.go`, checks that route patterns do not conflict under Go `http.ServeMux`, and compares the exact method/path set with this specification. Jenkins runs the same command.

The operation inventory is complete, but several write bodies and responses intentionally use generic JSON schemas. Domain-specific schemas and contract-conformance improvements remain follow-up work tracked in [the current task tracker](../CURRENT_TASKS.md) and [master documentation](../PROJECT_DOCUMENTATION.md).

To preview locally with Swagger UI:

```bash
docker run --rm -p 8088:8080 \
  -e SWAGGER_JSON=/spec/eomp-openapi-spec.yaml \
  -v "${PWD}/docs/openapi:/spec:ro" \
  swaggerapi/swagger-ui
```

Then open <http://localhost:8088>.
