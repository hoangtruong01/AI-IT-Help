module eomp/services/asset

go 1.25.0

require eomp/packages/shared v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/rabbitmq/amqp091-go v1.14.0 // indirect
	github.com/redis/go-redis/v9 v9.22.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace eomp/packages/shared => ../../packages/shared
