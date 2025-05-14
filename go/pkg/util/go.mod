module github.com/keybittech/awayto-v3/go/pkg/util

go 1.24.3

require (
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.23.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250425173222-7b384671a197
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250425153114-8976f5be98c1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ../handlers

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients
