Kurtosis CLI
============
This repo contains:
* The `kurtosis` CLI
* An internal testsuite to verify that the CLI (and Kurtosis) works

Developing
----------
* Run `scripts/build.sh` to build the CLI into a binary & testsuite into a Docker image
* Run `scripts/launch-cli.sh` to run arbitrary CLI commands with the locally-built binary
* Run `scripts/run-one-internal-testsuite.sh LANG` (replacing `LANG` with a language from the `supported-languages.txt` file) to run `kurtosis test` using the locally-built binary to run the internal testuite
* Run `scripts/run-all-internal-testsuites.sh` to run the internal testsuites in all languages

Debugging User Issues
---------------------
### The CLI's not working and there's not enough info to figure out why
The CLI has its own log level (separate from the engine, core, & modules). Set the `--cli-log-level` flag to `debug` or `trace` to see more info about what the CLI is doing (can be set on any command).

### Tab completion isn't working
Have the user run the following command, so that all the logs during completion get logged:

```
export BASH_COMP_DEBUG_FILE="/tmp/completion-debugging.log"
```

Cobra also ships with an invisible `__complete` command that will allow you to test various different scenarios like so (note that there needs to be an extra `""` at the point where the user is hitting tab!):

```
kurtosis __complete enclave inspect ""
```
