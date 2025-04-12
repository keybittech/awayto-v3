module github.com/keybittech/awayto-v3/go/pkg/util

go 1.24.0

require (
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/keybittech/awayto-v3/go/pkg/types v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.23.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250409194420-de1ac958c67a
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250307204501-0409229c3780.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
)

replace github.com/keybittech/awayto-v3/go/pkg/types => ../types

replace github.com/keybittech/awayto-v3/go/pkg/api => ../api

replace github.com/keybittech/awayto-v3/go/pkg/handlers => ../handlers

replace github.com/keybittech/awayto-v3/go/pkg/clients => ../clients
