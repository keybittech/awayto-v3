module github.com/keybittech/awayto-v3/go/pkg/handlers

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/keybittech/awayto-v3/go/pkg/clients v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/redis/go-redis/v9 v9.7.3 // indirect
	github.com/sashabaranov/go-openai v1.38.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/util => ../util

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api
