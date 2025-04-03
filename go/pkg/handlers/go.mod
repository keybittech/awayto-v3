module github.com/keybittech/awayto-v3/go/pkg/handlers

go 1.23.6

require (
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/keybittech/awayto-v3/go/pkg/interfaces v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.10.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/redis/go-redis/v9 v9.7.1 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/interfaces => ../interfaces

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/util => ../util

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api
