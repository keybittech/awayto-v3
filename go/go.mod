module github.com/keybittech/awayto-v3/go

go 1.24.0

require (
	github.com/joho/godotenv v1.5.1
	github.com/keybittech/awayto-v3/go/pkg/api v0.0.0-20250413170509-b98baad0beed
	github.com/keybittech/awayto-v3/go/pkg/clients v0.0.0-20250413170509-b98baad0beed
	github.com/keybittech/awayto-v3/go/pkg/handlers v0.0.0-20250413170509-b98baad0beed
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-20250413170509-b98baad0beed
	github.com/playwright-community/playwright-go v0.5101.0
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250307204501-0409229c3780.1 // indirect
	cel.dev/expr v0.19.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/bufbuild/protovalidate-go v0.9.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/redis/go-redis/v9 v9.7.3 // indirect
	github.com/sashabaranov/go-openai v1.38.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250409194420-de1ac958c67a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250404141209-ee84b53bf3d0 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/api => ./pkg/api

replace github.com/keybittech/awayto-v3/go/pkg/clients => ./pkg/clients

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ./pkg/handlers

replace github.com/keybittech/awayto-v3/go/pkg/util => ./pkg/util

replace github.com/keybittech/awayto-v3/go/pkg/types => ./pkg/types
