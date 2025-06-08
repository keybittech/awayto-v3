## Project: Awayto Exchange
Awayto is for groups or organizations which require scheduling, communications, reporting, and related functionality. Collaborative multi-modal features (voice/video/chat/docs) are positioned within a robust collection of operations-driven functionality (users/groups/roles/services/scheduling/surveying/reporting). Many founding principles of the platform were derived from previous experience developing and extending an online writing center. However, Awayto is built to be generic-purpose and enable the development of any kind of online platform with multi-modal needs. This is a unified server based on Go and Protobufs, with some generators to help build pre-determined mutexe-supported structs. As well, the frontend uses a revised implementation of RTK-Query in order to streamline the development and usage of API endpoints: defined Protobufs are used to generate an OpenAPI spec document, which in turn is used to auto-generate RTK-Query hooks for React. We follow these conventions to maintain consistency:

### Test Commands
Run a full integration test which starts up the server itself, creates a new group with some users and performs end to end tests across the application by making typical http requests from a standard go http client.
- make go_test_integration
Run benchmarks for various application metrics.
- make go_test_bench
Test the various go package modules.
- make go_test_unit_api
- make go_test_unit_clients
- make go_test_unit_handlers
- make go_test_unit_util

### Software Usage
In no particular order, the following lists the third-party software used in Awayto, along with their key features and a primary source for usage in the system, which could be a folder of items or a single file:
Technology 	Description 	Source
Make 	Task running, building, deploying 	Makefile
Shell 	Deployment install/configure scripts 	/deploy/scripts
Docker 	Container service, docker compose, supports cloud deployments 	/deploy/scripts/docker
Postgres 	Primary database 	/deploy/scripts/db
React 	Front end TSX components and hooks built with a customized CRACO config 	/ts
ReduxJS Toolkit 	React state management and API integrating with Protobufs 	/ts/src/hooks/store.ts
PNPM 	Front end package management 	/ts/package.json
Let’s Encrypt 	External certificate authority 	
Hetzner 	Cloud deployment variant 	/deploy/scripts/host
Keycloak 	Authentication and authorization, SSO, SAML, RBAC 	/java
Redis 	Sessions & caching 	/go/pkg/clients/redis.go
Hugo 	Static site generator for landing, documentation, marketing 	/landing
DayJS 	Scheduling and time management utilities 	/ts/src/hooks/time_unit.ts
Material-UI 	React UI framework based on Material Design 	/ts/src/modules
Coturn 	TURN & STUN server for WebRTC based voice and video calling 	/deploy/scripts/turn
WebSockets 	Dedicated websocket server for messaging orchestration, interactive whiteboard 	/go/pkg/clients/sock.go

### TSX Code Style
- Utilize "awayto-hooks" imports, which provide all the functionality in ts/src/hooks
- All data must be fetched using the variable siteApi which comes from awayto-hooks
- siteApi is populated by converting the service definitions in top-level protos folder into ts/openapi.yaml, which we then use rtk-query tools to convert the openapi.yaml to ts/hooks/api.ts -- this is an automated process therefore we should only seek to utilize siteApi and its methods, and never attempt to change it
- Only directly import mui components at their third level (see example), never use deconstruction/braces to import MUI components or images
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
- The "landing" is what users see when they visit the index of the website
- Hugo is used to do static-site generation based on markdown files
- Only focus on the content as if we were editing a marketing page
- Do not add any extra Javascript in the landing folder as it should be used extremely sparingly

### What NOT to Do
- Do not change code unrelated to the specific inquiry being asked.
- Do not comment your changes unless specifically asked.

### Now what?
After this colon the user has asked a specific question to do or investigate which is what we now need to do:
