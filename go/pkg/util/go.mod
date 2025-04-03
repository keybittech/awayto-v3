module github.com/keybittech/awayto-v3/go/pkg/util

go 1.23.6

require (
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.22.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb
	google.golang.org/protobuf v1.36.5
)

require buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ../handlers

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients

replace github.com/keybittech/awayto-v3/go/pkg/interfaces => ../interfaces
