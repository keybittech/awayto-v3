module github.com/keybittech/awayto-v3/go/pkg/clients

go 1.24.0

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-20250413170509-b98baad0beed
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.7.3
	github.com/sashabaranov/go-openai v1.38.0
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250425153114-8976f5be98c1.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250425173222-7b384671a197 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/util => ../util

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ../handlers

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api
