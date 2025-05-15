module github.com/keybittech/awayto-v3/go

go 1.24.3

require (
	github.com/joho/godotenv v1.5.1
	github.com/keybittech/awayto-v3/go/pkg/api v0.0.0-20250515042730-535705ff7e23
	github.com/keybittech/awayto-v3/go/pkg/clients v0.0.0-20250515042730-535705ff7e23
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-20250515042730-535705ff7e23
	github.com/playwright-community/playwright-go v0.5101.0
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6
	golang.org/x/time v0.11.0
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250425153114-8976f5be98c1.1 // indirect
	cel.dev/expr v0.24.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/bufbuild/protovalidate-go v0.10.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/deckarep/golang-set/v2 v2.8.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.25.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.4 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/keybittech/awayto-v3/go/pkg/handlers v0.0.0-20250515042730-535705ff7e23 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/redis/go-redis/v9 v9.8.0 // indirect
	github.com/sashabaranov/go-openai v1.40.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250512202823-5a2f75b736a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250512202823-5a2f75b736a9 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/api => ./pkg/api

replace github.com/keybittech/awayto-v3/go/pkg/clients => ./pkg/clients

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ./pkg/handlers

replace github.com/keybittech/awayto-v3/go/pkg/util => ./pkg/util

replace github.com/keybittech/awayto-v3/go/pkg/types => ./pkg/types
