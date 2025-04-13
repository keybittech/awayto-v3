module github.com/keybittech/awayto-v3/go/pkg/api

go 1.24.0

require (
	github.com/bufbuild/protovalidate-go v0.9.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.7.3
	golang.org/x/time v0.11.0
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250307204501-0409229c3780.1 // indirect
	cel.dev/expr v0.19.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/keybittech/awayto-v3/go/pkg/clients v0.0.0-20250413170509-b98baad0beed // indirect
	github.com/keybittech/awayto-v3/go/pkg/handlers v0.0.0-20250413170509-b98baad0beed // indirect
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000 // indirect
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-20250413170509-b98baad0beed // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/sashabaranov/go-openai v1.38.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250409194420-de1ac958c67a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250404141209-ee84b53bf3d0 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/util => ../util

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ../handlers

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients
