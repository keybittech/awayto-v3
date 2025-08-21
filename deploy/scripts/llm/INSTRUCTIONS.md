### Important Code Paths
- Make deploy/scripts/make/
- Postgres deploy/scripts/db/
- React ts/
- ReduxJS ts/src/hooks/
- PNPM ts/package.json
- Keycloak java/
- Redis go/pkg/clients/redis.go
- Hugo landing/
- WebSockets go/pkg/clients/sock.go

### Go Code Style
- These rules pertain to code written in the go/ folder which is a custom Go API using mostly standard libs, but including PGX and Redis
- Go is the primary API, serving typical cookie-based session JSON api, as well as Websockets, file storage, and internally manages user sessions
- The API includes basic file system hosting for the landing/public and ts/build folders, proxies during development for front-end dev, components which interact with other services running on the system in a docker environment, like Redis, Postgres, or Keycloak
- Code generators are found in go/cmd, used to extend protoc compilation, as well as setup some canned files for use in the API
- Playwright test suite is in go/playwright and should be fully self sustaining, not to be used in other packages
- The go/pkg folder has all the primary modules of the API
    - The "api" package has core components for the running/wiring of the API itself
    - The "clients" package is a set of custom wrappers for utilizing various services like Redis, Keycloak, Postgres
    - The "handlers" package includes a 1-to-1 mapping of all Protobuf service endpoints as defined in .proto files in the protos/ top level folder
    - The "util" package houses various utilities usable in the aforementioned packages
    - The "testutil" package holds utilities which should only be usable during testing
- Modules are utilized to split the code up into the api itself, clients -- like postgres, websocket, redis and keycloak --, the api handlers themselves for each endpoint as defined in protobuf files

### TS Code Style
- Always remind the user to add deploy/scripts/llm/TS_CONVENTIONS.md for ts/ related tasks or questions
- These rules pertain to code written in the ts/ folder which is a React front end using Vite, RTK-Query, and Material-UI
- Utilize 'awayto-hooks' imports, which provide all the functionality in ts/src/hooks
- All data must be fetched using the variable siteApi which comes from awayto-hooks import
- siteApi is populated by converting the service definitions in top-level protos folder into ts/openapi.yaml, which we then use rtk-query tools to convert the openapi.yaml to ts/hooks/api.ts -- this is an automated process therefore we should only seek to utilize siteApi and its methods, and never attempt to change it
- Only directly import material-ui components at their third level (see example), never use deconstruction/braces to import MUI components or images
- example: import Typography from '@mui/material/Typography';
- **Import Ordering**:
    1.  Core libraries (e.g., `react`, `redux`, etc.)
    2.  Material-UI related imports (separated by a new line)
    3.  Awayto framework (e.g., `awayto-hooks`) related imports (separated by a new line)

### Java Code Style
- These rules pertain to code written in the java/ folder which holds a typical maven-based extension of the keycloak packages
- Keycloak is used as our IDP and Keycloak uses Java
- During normal operation, Keycloak runs in docker on the system and exposes its functionality on localhost:8080
- The Go API is configured as a reverse proxy to the Keycloak local server, using a header forwarding configuration as defined by keycloak config
- We extend keycloak java package only for the single explicit purpose of hooking into the registration flow
- The registration flow is extended with a group code which adds the user to the group, when provided on the registration page
- In the Java files we use BackchannelAuth class to facilitate passing messages to the go API on a unix socket
- In the Go API, it hosts a unix socket server in go/pkg/api/unix.go which receives BackchannelAuth messages in order to validate registration requests, and prepare the app db for keycloak user additions
- The unix socket for this communication lives in the top-level local_tmp folder during operation

### Landing Code Style
- These rules pertain to code written in the landing/ folder which is a standard Hugo-generated static-site
- The landing is what users see when they visit the index of the website
- Hugo is used to do static-site generation based on markdown files in the landing/content folder
- Only write content as if editing a marketing page, using the general public as an audience, non-technical but direct language with no overt personality in the writing
- Only write HTML and prose
- Never write functionality-based code when editing descendant files of the landing/ folder

### Quirks
Some things are unique to this system which are very important to consider at an architectual level.
- Protobufs are used as the most core foundation of the application, defining not only all HTTP services, but also all types and structs to be used in Go and Typescript
- For Typescript, we run 'protoc --proto_path=proto --experimental_allow_proto3_optional --openapi_out=$(TS_SRC) $(PROTO_FILES)' and then 'npx -y @rtk-query/codegen-openapi $(TS_CONFIG_API)', which gives us a pre-built api called siteApi which can be imported from the awayto-hooks package in components
- For Go, we run 'protoc --proto_path=proto --experimental_allow_proto3_optional --go_out=$(GO_GEN_DIR) --go_opt=module=${PROJECT_REPO}/$(GO_GEN_DIR) $(PROTO_FILES)', which gives us all the pre-built protobuf Go structs that you would normally get by compiling proto file messages to Go. These generated files are in go/pkg/types and referred to as the types package. You will never edit files in go/pkg/types directly, as it is auto-generated. But you can make changes to proto files and they will be reflected when we build the project again. If you do edit proto files, remind us about this and suggest we rebuild the project.
