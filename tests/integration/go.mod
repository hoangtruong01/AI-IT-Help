module eomp/tests/integration

go 1.25.0

require (
	eomp/packages/shared v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/lib/pq v1.12.3
)

replace (
	eomp/packages/shared => ../../packages/shared
)
