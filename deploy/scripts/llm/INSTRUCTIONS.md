## Project: Awayto Exchange
Awayto is for groups or organizations which require scheduling, communications, reporting, and related functionality. Collaborative multi-modal features (voice/video/chat/docs) are positioned within a robust collection of operations-driven functionality (users/groups/roles/services/scheduling/surveying/reporting). Many founding principles of the platform were derived from previous experience developing and extending an online writing center. However, Awayto is built to be generic-purpose and enable the development of any kind of online platform with multi-modal needs.

You have access to the complete filesystem, at least as much is needed for you to perform your work. The overall premise of the application is a Go lang server which supports the functionalities listed above. Review the list of software to understand what and where important modules are.This is a unified server based on Go and Protobufs, with some generators to help build pre-determined mutexe-supported structs. As well, the frontend uses a revised implementation of RTK-Query in order to streamline the development and usage of API endpoints: defined Protobufs are used to generate an OpenAPI spec document, which in turn is used to auto-generate RTK-Query hooks for React.

### Working Procedure
The server is already running in watch mode for folders go/ and ts/. Complete the task then run one of these make commands to verify things are working:
- make go_test_unit
- make go_test_ui

### Encountering Read-Only
Some files are marked as read only. You may change them, but only when the following conditions are met:
- Identify the read only file you are requesting to change
- Determine what content needs to be changed in natural language in conversation to us
- Consider your task complete in full, and end all future operatives; when we review your request, that work will be performed in the future

### Software Locations:
- Make /workspace/deploy/scripts/make
- Postgres /workspace/deploy/scripts/db
- React /workspace/ts
- ReduxJS /workspace/ts/src/hooks
- PNPM /workspace/ts/package.json
- Keycloak /workspace/java
- Redis /workspace/go/pkg/clients/redis.go
- Hugo /workspace/landing
- WebSockets /workspace/go/pkg/clients/sock.go

### TSX Code Style
- Utilize 'awayto-hooks' imports, which provide all the functionality in ts/src/hooks
- All data must be fetched using the variable siteApi which comes from awayto-hooks import
- siteApi is populated by converting the service definitions in top-level protos folder into ts/openapi.yaml, which we then use rtk-query tools to convert the openapi.yaml to ts/hooks/api.ts -- this is an automated process therefore we should only seek to utilize siteApi and its methods, and never attempt to change it
- Only directly import material-ui components at their third level (see example), never use deconstruction/braces to import MUI components or images
- example: import Typography from '@mui/material/Typography';

### Go Code Style
- Go is the primary API, serving typical cookie-based session JSON api, as well as Websockets, file storage, and internally manages user sessions by communicating with keycloak internally on keycloak's admin api running on localhost:8080 -- not publically accessible.
- Modules are utilized to split the code up into the api itself, clients -- like postgres, websocket, redis and keycloak --, the api handlers themselves for each endpoint as defined in protobuf files, 

### Java Code Style
- Keycloak is used as our IDP and Keycloak uses Java
- We extend keycloak only for the single explicit purpose of hooking into the registration flow
- The registration flow is extended with a group code which adds the user to the group, when provided
- Keycloak and Go server interact through BackchannelAuth methods, which is a unix socket on the system in local_tmp folder

### Landing Code Style
- The landing is what users see when they visit the index of the website
- Hugo is used to do static-site generation based on markdown files
- Only focus on the content as if we were editing a marketing page
- Do not add any extra Javascript in the landing folder as it should be used extremely sparingly

### Quirks
Some things are unique to this system which are very important to consider at an architectual level.
- Protobufs are used as the most core foundation of the application, defining not only all HTTP services, but also all types and structs to be used in Go and Typescript
- For Typescript, we run 'protoc --proto_path=proto --experimental_allow_proto3_optional --openapi_out=$(TS_SRC) $(PROTO_FILES)' and then 'npx -y @rtk-query/codegen-openapi $(TS_CONFIG_API)', which gives us a pre-built api called siteApi which can be imported from the awayto-hooks package in components
- For Go, we run 'protoc --proto_path=proto --experimental_allow_proto3_optional --go_out=$(GO_GEN_DIR) --go_opt=module=${PROJECT_REPO}/$(GO_GEN_DIR) $(PROTO_FILES)', which gives us all the pre-built protobuf Go structs that you would normally get by compiling proto file messages to Go. These generated files are in go/pkg/types and referred to as the types package.
