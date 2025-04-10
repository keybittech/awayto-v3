module github.com/keybittech/awayto-v3/go

go 1.24.0

require (
	github.com/joho/godotenv v1.5.1
	github.com/keybittech/awayto-v3/go/pkg/api v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/clients v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/handlers v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/interfaces v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-00010101000000-000000000000
	github.com/playwright-community/playwright-go v0.4902.0
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8
	google.golang.org/protobuf v1.36.5
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect
	cel.dev/expr v0.19.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/bufbuild/protovalidate-go v0.9.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set/v2 v2.7.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/redis/go-redis/v9 v9.7.3 // indirect
	github.com/sashabaranov/go-openai v1.38.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250227231956-55c901821b1e // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/api => ./pkg/api

replace github.com/keybittech/awayto-v3/go/pkg/clients => ./pkg/clients

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ./pkg/handlers

replace github.com/keybittech/awayto-v3/go/pkg/util => ./pkg/util

replace github.com/keybittech/awayto-v3/go/pkg/types => ./pkg/types

replace github.com/keybittech/awayto-v3/go/pkg/interfaces => ./pkg/interfaces
