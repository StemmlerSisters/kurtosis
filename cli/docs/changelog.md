# TBD

### Features

* Added TypeScript test: `network_partition.test.ts`
* Added TypeScript test: `files.test.ts`
* Added TypeScript test: `module.test.ts`
* Added TypeScript test: `wait_for_endpoint_availability.test.ts`

# 0.8.6
### Features
* Upgrade to engine server 1.8.2 which adds the removal of dangling folders in the clean endpoint


# 0.8.5
### Features
* Upgrade to engine server 1.8.1 which adds `Kurtosis Engine` checker in `KurtosisContext` creation

### Features
* Added TypeScript test: `files_artifact_mounting.test.ts`
* Added TypeScript test: `exec_command.test.ts`

# 0.8.4
### Features
* The `module exec` command prints and follows the module's logs

### Changes
* Upgraded to the following dependencies to support ID label:
  * obj-attrs-schema-lib -> 0.6.0
  * core dependencies -> 1.36.11
  * engine dependencies -> 1.7.7

# 0.8.3
### Features
* Added call to clean endpoint


# 0.8.2
### Fixes
* Make a best-effort attempt to pull module & user service images before starting them
* The headers of `enclave inspect` are printed in deterministic, sorted order so they stop jumping around on subsequent runs
* Fix a bug where the `sandbox` would error on `enclaveCtx.getServiceContext`

# 0.8.1
### Features
* Added TypeScript test: `bulk_command_execution.test.ts`
* Added TypeScript test: `basic_datastore_and_api.test.ts`

### Fixes
* Add proper logging for `enclave_setup.ts`
* Fixed a bug with the Javascript REPL image
* Resolve TypeScript test minor bug
* Upgrade to engine server 1.7.3
* Upgrade Typescript to the latest version of the engine-server library

# 0.8.0
### Features
* Upgraded to the following dependencies to support users specifying a user-friendly port ID for their ports:
    * obj-attrs-schema-lib -> 0.5.0
    * core dependencies -> 1.36.9
    * engine dependencies -> 1.7.2
* Added `network_soft_partition_test` in golang internal test suite
* Added a unit test to ensure that an API break in the engine (which will require restarting the engine) is an API break for the CLI

### Fixes
* When the engine server API version that the CLI expects doesn't match the running engine server's API version, the user gets an error and is forced to restart their engine

### Breaking Changes
* Upgraded the engine server to 1.7.2
    * Users will need to run `kurtosis engine restart` after upgrading to this version of the CLI
    * Engine API users (e.g. in tests) will need to update to `kurtosis-engine-api-lib` 1.7.2
    * Module users will need to update their modules to [Module API Lib](https://github.com/kurtosis-tech/kurtosis-module-api-lib) 0.12.3

# 0.7.4
### Features
* Added a `--partitioning` flag to `module exec` for enabling partitioning

### Changes
* The Go internal testsuite's enclaves will now be named with Unix millis, rather than Unix seconds
* Partitioning defaults to false for `module exec`

# 0.7.3
### Features
* Added new command named service shell (kurtosis service shell enclave_id service_id) which performs the same as docker exec -it container_id sh

# 0.7.2
### Features
* The enclave for `module exec` will now be named after the module image and the time it was ran
* Allow users running `module exec` to manually specify the ID of the enclave that will be created

### Fixes
* Fixed a bug where `service logs` that was successful would really fail

# 0.7.1
### Fixes
* Attempt to fix CLI artifact publishing

# 0.7.0
### Features
* Added TypeScript: `basic_datastore_test.ts`, `enclave_setup.ts` and `test_helpers.ts`
* Added TypeScript project inside of `internal_testsuites` folder.
* The `test_helpers` class now has a higher-level API: `AddDatastoreService` and `AddAPIService`, which makes many of our internal testsuite test setups a one-liner
* Add an extra API container status result to `enclave inspect`
* Reimplement endpoint availability-waiting test in new Go test framework

### Fixes
* `stacktrace.Propagate` now panics when it gets a `nil` value
* Fixed bug in files artifact mounting test where it would fail on Mac (but not Linux)
* Fixed `enclave inspect` breaking if the enclave was stopped
* Fixed publishing, which was temporarily broken

### Changes
* The Javascript REPL now uses Node 16.13.0 (up from 16.7.0)
* Grouped all the internal testsuites into a single directory
* Gave `build.sh` scripts to the CLI & internal testsuite subdirectories now
* There is no longer a root `go.mod`, but now one in CLI and one in `golang_internal_testsuite` (rationale being that the dependencies for the CLI and for the internal testsuite are very different, plus we'll have a `typescript_internal_testsuite` soon)
* Removed the "local static" element to `localStaticFilesTest`, because there's no longer a distinction between "local" and "static" now that the testsuite runs with Go test
* The `--image` arg to `engine start` and `engine restart` has been replaced with a `--version` flag, so that the full org & image is no longer required
* The `--kurtosis-api-image` flag to `sandbox` has been replaced with a `--api-container-version` flag, so that the full org & image is no longer required
* The `--api-container-image` flag to `enclave new` has been replaced with a `--api-container-version` flag, so that the full org & image is no longer required
* The `--api-container-image` flag to `module exec` has been replaced with a `--api-container-version` flag, so that the full org & image is no longer required
* The `engine status` now returns the engine version, rather than the API version
* Use engine-server 1.5.6
* Remove all references to Palantir stacktrace

### Removals
* Removed the `test` command, as tests can be written directly in your testing framework of choice by connecting to the running engine using `kurtosis-engine-api-lib`
* Removed the `AdvancedNetworkTest`, because we no longer have `Network` types
* Removed alternating colors for the tables, because it's a pain to maintain
* Stop publishing a `latest` version of the REPL images, because the CLI should use pinned X.Y.Z version to avoid problems

### Breaking Changes
* Removed the `test` command
    * Users should migrate their tests out of the Kurtosis testing framework, and into a testing framework of choice in their language
* The `--image` arg to `engine start` and `engine restart` has been replaced with a `--version` flag
    * Users should use the new flag with the Docker tag of the engine to start
* The `--kurtosis-api-image` flag to `sandbox` has been replaced with a `--api-container-version` flag
    * Users should use the new flag with the Docker tag of the API container to start
* The `--api-container-image` flag to `enclave new` has been replaced with a `--api-container-version` flag
    * Users should use the new flag with the Docker tag of the API container to start
* The `--api-container-image` flag to `module exec` has been replaced with a `--api-container-version` flag
    * Users should use the new flag with the Docker tag of the API container to start

# 0.6.8
### Features
* Added a cache file for getting the latest released CLI version from GitHUB API

# 0.6.7
### Features
* Added Yellow and White alternating colors in TablePrinter

### Fixes
* The `kurtosis engine restart` suggestion when the engine is out-of-date now:
    * No longer has a trailing space
    * Is on the same line as the "engine is out-of-date" message

# 0.6.6
### Features
* The API Container host port was added when showing the data with the command `enclave inspect`

# 0.6.5
### Fixes
* `enclave inspect` host port bindings now properly return `127.0.0.1`, to match what's returned by the `AddService`

# 0.6.4
### Fixes
* Fixed issue #69 - now the CLI version checker passes when the current version is newer than the latest public version (e.g. during a release)
* Enable the unit test for the `RootCmd` because issue #69 is fixed

### Changes
* Removed `version_checker` class, now the `checkCLILatestVersion` functionality is part of the `RootCmd` and the `checkIfEngineIsUpToDate` functionality is controlled by the `engine_existence_guarantor` 

# 0.6.3
### Fixes
* Temporarily disable the unit test for the `RootCmd` until issue #69 is fixed

# 0.6.2
### Features
* `enclave inspect` also prints the `Kurtosis modules`
* Added `version_checker.CheckLatestVersion` method to check if it is running the latest CLI version before running any CLI command
* Added `version_checker.CheckIfEngineIsUpToDate` method to check if it is running engine is up-to-date.

### Fixes
* Fixed error where `service logs` command is executed with a nonexistent enclave ID or nonexistent GUID just silently exits
* Upgraded to engine server 0.5.2, which returns host port bindings in the format `127.0.0.1` rather than `0.0.0.0` for Windows users
* Upped the run timeouts of the advanced network test, module test, and bulk command execution test to 90s

### Changes
* Replaced `kurtosistech/example-microservices_datastore` Docker image with `kurtosistech/example-datastore-server` in `golang_internal_testsuite`
* Replaced `kurtosistech/example-microservices_api` Docker image with `kurtosistech/example-api-server` in `golang_internal_testsuite`
* Replaced `kurtosistech/example-microservices_datastore` Docker image with `docker/getting-started` in `bulk_command_execution_test` and `wait_for_endpoint_availability_test`
* Upgraded `datastore army module` Docker image to the latest version `kurtosistech/datastore-army-module:0.1.5` in `module_test` 

# 0.6.1
### Fixes
* Fixed a bug where a testsuite could be reported as passing, even when the tests were failing

# 0.6.0
### Features
* The engine loglevel can be configured with the new `--log-level` flag to `engine start`

### Changes
* Renamed the `repl new` image flag to make more sense
* Update to using the engine server that stores engine/enclave data on the user's local machine

### Breaking Changes
* The `--js-repl-image` flag of `repl new` has been renamed to `--image`, with shorthand `-i`

# 0.5.5
### Fixes
* A failed `module exec` or `sandbox` stops, rather than destroys, the enclave it created

# 0.5.4
### Features
* `enclave new` prints the new enclave's ID
* Information about how to stop or remove the enclave created by `sandbox` is printed after the REPL exits
* Added a `clean` command, to clean up accumulated Kurtosis artifacts
* Added a `repl inspect` command to list installed packages on the REPL

### Fixes
* Use `--image` flag in `kurtosis engine start` command, it was not being used when the engine is being executed
* Fix the returning values order when `DestroyEnclave` method is called in `kurtosis sandbox` command
* Fixed a bug where `engine status` wouldn't check the error value from getting the engine status object
* The Javascript REPL's module installation paths in the Dockerfile are now filled from Go code constants (rather than being hardcoded)

### Changes
* The `sandbox` command no longer destroys the enclave after the REPL exits
* Upgrade to engine server 0.4.7, where API container doesn't shut down the containers after it exits (instead relying on the engine server to do that)

# 0.5.3
### Fixes
* Upgrade to the `goreleaser-ci-image` 0.1.1 to publish a new Homebrew formula with a fix for the `bottle :unneeded` deprecation warning

# 0.5.2
### Features
* Added `enclave stop` and `enclave rm` commands

# 0.5.1
### Features
* Add instructions for users on what to do if no Kurtosis engine is running
* If an engine isn't running, the CLI will try to start one automatically

# 0.5.0
### Changes
* Replaced `EnclaveManager` with `Kurtosis Engine API Libs` which handle all the interactions with the `Kurtosis Engine Server`

### Features
* Add a `version` command to print the CLI's version, with a test
* Added a global `--cli-log-level` flag that controls what level the CLI will log at

### Fixes
* The Kurtosis Client version used by the JS REPL image will now use the `KurtosisApiVersion` constant published by Kurt Client
* Fixed bug where testsuite containers weren't getting any labels

### Breaking Changes
* Interactions with the CLI now require a Kurtosis engine to be running
    * Users should run `kurtosis engine start` to start an engine

# 0.4.3
### Features
* Added documentation in README about how to develop on this repo
* Upgraded to `kurtosis-core` 1.25.2, which contains fixes for `container-engine-lib` 0.7.0 feature that allows binding container ports to specific host machine ports
* Added `engine start` command to the CLI
* Added `engine stop` command to the CLI
* `engine start` waits until the engine is responding to gRPC requests before it reports the engine as up
* Added `engine status` command to the CLI
* Start a Kurtosis engine server in the CI environment

### Fixes
* Clarified the difference between the two types of params in `module exec`
* `engine start` won't start another container if one is already running
* `engine start` waits for gRPC availability before it reports the engine up

# 0.4.2
### Features
* `enclave` commands also show enclave state
* Standardized table-printing logic into a `TablePrinter` object 
* Added a `KeyValuePrinter` for pretty-printing key-value pairs
* `enclave inspect` also prints the enclave ID & state

### Fixes
* `module exec` will attempt to update the module & API container images before running
* Fixed a bug where having a `node_modules` directory in your current directory when starting a REPL will cause the REPL to fail

### Changes
* Upgrade to testsuite-api-lib 0.11.0, which uses Kurt Client 0.19.0 (already handled in v0.4.0 of this repo)
* When running a REPL, your current directory is now mounted at `/local` rather than the present directory

# 0.4.1
### Fixes
* Update the Javascript CLI's `core-api-lib` version to 0.19.0, to match the rest of the code

# 0.4.0
### Changes
* Switched all references to "Lambda" to "module"

### Fixes
* `ModuleTest` (renamed from `LambdaTest`) now uses the ports returned by the Datastore Army module
* Fixed bug in CI where `pipefail` wasn't set which would result in the testsuite-running step passing when it shouldn't

### Breaking Changes
* Renamed the `lambda` command to `module`

# 0.3.4
### Fixes
* Stop attempting to upload APK packages to Gemfury (which can't accept APK packages and throws an error)

# 0.3.3
### Features
* Added a `repl install` command for installing NPM packages to a running REPL container
* `ParsePositionalArgs` (renamed to `ParsePositionalArgsAndRejectEmptyStrings`) now also errors on empty strings

# 0.3.2
### Features
* Added `enclave new` command to create a new enclave
* Added `repl new` command to start a new Javascript REPL
* Added `REPL runner` to reuse the creation and execution of the REPL container
* Print interactive REPLs in `enclave inspect`
* Added `GetEnclave` method in `EnclaveManager` in order to get information of a running enclave
* Upgrade Kurtosis Core Engine Libs to v0.6.0 which adds `Network` type
* Upgrade Kurtosis Core to v1.24.0 which splits `api-container-url` into `api-container-ip` and `api-container-port`

# 0.3.1
### Fixes
* Pinned the default API container version to the same version as in the `go.mod`, so that its version can't silently upgrade under users and break everything

# 0.3.0
### Changes
* Changed the Homebrew/deb/rpm package name to `kurtosis-cli` (was `kurtosis`)

### Breaking Changes
* The CLI is now installed via the `kurtosis-cli` package (for Homebrew, APT, and Yum) rather than just `kurtosis`

# 0.2.1
### Fixes
* Fixed missing `FURY_TOKEN` when publishing

# 0.2.0
### Features
* Ported over the CLI & internal testsuite from `kurtosis-core`

### Breaking Changes
* Changed pretty much everything to add the CLI

# 0.1.0
* Initial commit
