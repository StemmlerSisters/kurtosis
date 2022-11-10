# TBD
### Breaking Changes
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` engine's endpoints, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `GetUserServiceLogsArgs` engine's proto file
  - Users should have to add this param in all the `GetUserServiceLogs` calls
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` methods in `KurtosisContext`, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `KurtosisContext.GetUserServiceLogs`
  - Users will have to addition this new parameter on every call

### Changes
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON
- Moved `read_file` to be a simple Startosis builtin in place of a Kurtosis instruction

# 0.51.13
### Fixes
- Set `entry_point_args` and `cmd_args` to `nil` if not specified instead of empty array 

# 0.51.12
### Features
- Added an optional `--dry-run` flag to the `startosis exec` (defaulting to false) command which prints the list of Kurtosis instruction without executing any. When `--dry-run` is set to false, the list of Kurtosis instructions is printed to the output of CLI after being executed.

# 0.51.11
### Features
- Improve how kurtosis instructions are canonicalized with a universal canonicalizer. Each instruction is now printed on multiple lines with a comment pointing the to position in the source code.
- Support `private_ip_address_placeholder` to be passed in `service_config` for `add_service` in Starlark

### Changes
- Updated how we generate the canonical string for Kurtosis `upload_files` instruction

# 0.51.10
### Changes
- Added Starlark `proto` module, such that you can now do `proto.has(msg, "field_name")` in Startosis to differentiate between when a field is set to its default value and when it is unset (the field has to be marked as optional) in the proto file though.

# 0.51.9
### Features
- Implemented the new `StreamUserServiceLogs` endpoint in the Kurtosis engine server
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Golang library
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Typescript library
- Added the `StreamUserServiceLogs` method in Loki logs database client
- Added `stream-logs` test in Golang and Typescript `internal-testsuites`
- Added `service.GUID` field in `Service.Ctx` in the Kurtosis SDK

### Changes
- Updated the CLI `service logs` command in order to use the new `KurtosisContext.StreamUserServiceLogs` when user requested to follow logs
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON

# 0.51.8
### Features
- Added exec and HTTP request facts
- Prints out the instruction line, col & filename in the execution error
- Prints out the instruction line, col & filename in the validation error
- Added `remove_service` to Startosis

### Fixes
- Fixed nil accesses on Fact Engine

### Changes
- Add more integration tests for Kurtosis modules with input and output types

# 0.51.7
### Fixes
- Fixed instruction position to work with nested functions

### Features
- Instruction position now contains the filename too

# 0.51.6
### Features
- Added an `import_types` Starlark instruction to read types from a .proto file inside a module
- Added the `time` module for Starlark to the interpreter
- Added the ability for a Starlark module to take input args when a `ModuleInput` in the module `types.proto` file

# 0.51.5
### Fixes
- Testsuite CircleCI jobs also short-circuit if the only changes are to docs, to prevent them failing due to no CLI artifact

# 0.51.4
### Fixes
- Fixed a bug in `GetLogsCollector` that was failing when there is an old logs collector container running that doesn't publish the TCP port
- Add missing bindings to Kubernetes gateway

### Changes
- Adding/removing methods from `.proto` files will now be compile errors in Go code, rather than failing at runtime
- Consolidated the core & engine Protobuf regeneration scripts into a single one

### Features
- Validate service IDs on Startosis commands

# 0.51.3
### Fixes
- Added `protoc` install step to the `publish_api_container_server_image` CircleCI task

# 0.51.2
### Features
- Added a `render_templates` command to Startosis
- Implemented backend for facts engine
- Added a `proto_file_store` in charge of compiling Startosis module's .proto file on the fly and storing their FileDescriptorSet in memory

### Changes
- Simplified own-version constant generation by checking in `kurtosis_version` directory

# 0.51.1
- Added an `exec` command to Startosis
- Added a `store_files_from_service` command to Startosis
- Added the ability to pass `files_artifact_mount_dirpaths` to the service config
- Added a `read_file` command to Startosis
- Added the ability to execute local modules in Startosis

### Changes
- Fixed a typo in a filename

### Fixes
- Fixed a bug in exec where we'd propagate a `nil` error
- Made the `startosis_module_test` in js & golang deterministic and avoid race conditions during parallel runs

### Removals
- Removed  stale `scripts/run-pre-release-scripts` which isn't used anywhere and is invalid.

# 0.51.0
### Breaking Changes
- Updated `kurtosisBackend.CreateLogsCollector` method in `container-engine-lib`, added the `logsCollectorTcpPortNumber` parameter
  - Users will need to update all the `kurtosisBackend.CreateLogsCollector` setting the logs collector `TCP` port number 

### Features
- Added `KurtosisContext.GetUserServiceLogs` method in `golang` and `typescript` api libraries
- Added the public documentation for the new `KurtosisContext.GetUserServiceLogs` method
- Added `GetUserServiceLogs` in Kurtosis engine gateway
- Implemented IP address references for services
- Added the `defaultTcpLogsCollectorPortNum` with `9713` value in `EngineManager`
- Added the `LogsCollectorAvailabilityChecker` interface

### Changes
- Add back old enclave continuity test
- Updated the `FluentbitAvailabilityChecker` constructor now it also receives the IP address as a parameter instead of using `localhost`
- Published the `FluentbitAvailabilityChecker` constructor for using it during starting modules and user services
- Refactored `service logs` Kurtosis CLI command in order to get the user service logs from the `logs database` (implemented in Docker cluster so far)

# 0.50.2
### Fixes
- Fixes how the push cli artifacts & publish engine runs by generating kurtosis_version before hand

# 0.50.1
### Fixes
- Fix generate scripts to take passed version on release

# 0.50.0
### Features
- Created new engine's endpoint `GetUserServiceLogs` for consuming user service container logs from the logs database server
- Added `LogsDatabaseClient` interface for defining the behaviour for consuming logs from the centralized logs database
- Added `LokiLogsDatabaseClient` which implements `LogsDatabaseClient` for consuming logs from a Loki's server
- Added `KurtosisBackendLogsClient` which implements `LogsDatabaseClient` for consuming user service container logs using `KurtosisBackend`
- Created the `LogsDatabase` object in `container-engine-lib`
- Created the `LogsCollector` object in `container-engine-lib`
- Added `LogsDatabase` CRUD methods in `Docker` Kurtosis backend
- Added `LogsCollector` CRUD methods in `Docker` Kurtosis backend
- Added `ServiceNetwork` (interface), `DefaultServiceNetwork` and `MockServiceNetwork`

### Breaking Changes
- Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
    - Users will need to update all the `CreateEngine` calls removing this parameter
- Updated `NewEngineServerArgs`,  `LaunchWithDefaultVersion` and `LaunchWithCustomVersion` methods in `engine_server_launcher` removed the `logsCollectorHttpPortNumber` parameter
    - Users will need to update these method calls removing this parameter
  
### Changes
- Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`
- Made some changes to the implementation of the module manager based on some PR comments by Kevin

### Features
- Implement Startosis add_service image pull validation
- Startosis scripts can now be run from the CLI: `kurtosis startosis exec path/to/script/file --enclave-id <ENCLAVE_ID>`
- Implemented Startosis load method to load from Github repositories

### Fixes
- Fix IP address placeholder injected by default in Startosis instructions. It used to be empty, which is invalid now
it is set to `KURTOSIS_IP_ADDR_PLACEHOLDER`
- Fix enclave inspect CLI command error when there are additional port bindings
- Fix a stale message the run-all-test-against-latest-code script
- Fix bug that creates database while running local unit tests
- Manually truncate string instead of using `k8s.io/utils/strings`

### Removals
- Removes version constants within launchers and cli in favor of centralized generated version constant
- Removes remote-docker-setup from the `build_cli` job in Circle

# 0.49.9
### Features
- Implement Startosis add_service method
- Enable linter on Startosis codebase

# 0.49.8
### Changes
- Added a linter
- Made changes based on the linters output
- Made the `discord` command a LowLevelKurtosisCommand instead of an EngineConsumingKurtosisCommand

### Features
- API container now saves free IPs on a local database

### Fixes
- Fix go.mod for commons & cli to reflect monorepo and replaced imports with write package name
- Move linter core/server linter config to within core/server

# 0.49.7
### Features
- Implement skeleton for the Startosis engine

### Fixes
- Fixed a message that referred to an old repo.

### Changes
- Added `cli` to the monorepo

# 0.49.6
### Fixes
- Fixed a bug where engine launcher would try to launch older docker image `kurtosistech/kurtosis-engine-server`.

# 0.49.5
### Changes
- Added `kurtosis-engine-server` to the monorepo
- Merged the `kurtosis-engine-sdk` & `kurtosis-core-sdk`

### Removals
- Remove unused variables from Docker Kurtosis backend

# 0.49.4
### Fixes
- Fix historical changelog for `kurtosis-core`
- Don't check for grpc proxy to be available

# 0.49.3
### Fixes
- Fix typescript package releases

# 0.49.2
### Removals
- Remove envoy proxy from docker image. No envoy proxy is being run anymore, effectively removing HTTP1.

### Changes
- Added `kurtosis-core` to the monorepo

### Fixes
- Fixed circle to not docs check on merge

# 0.49.1
### Fixes
- Attempting to fix the release version
### Changes
- Added container-engine-lib

# 0.49.0
### Changes
- This version is a dummy version to set the minimum. We pick a version greater than the current version of the CLI (0.29.1). 
