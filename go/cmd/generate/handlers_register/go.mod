module github.com/keybittech/awayto-v3/go/cmd/generate/handlers_register

go 1.24.3

require github.com/keybittech/awayto-v3/go/pkg/util v0.0.0-20250515042730-535705ff7e23

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250425153114-8976f5be98c1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.4 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250512202823-5a2f75b736a9 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/types => ../../../pkg/types

replace github.com/keybittech/awayto-v3/go/pkg/util => ../../../pkg/util
