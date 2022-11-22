#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
PARALLELISM=2
TIMEOUT="4m"   # This must be Go-parseable timeout

TESTSUITE_CLUSTER_BACKEND_DOCKER="docker"
TESTSUITE_CLUSTER_BACKEND_MINIKUBE="minikube"

# By default, run testsuite against docker
DEFAULT_TESTSUITE_CLUSTER_BACKEND="${TESTSUITE_CLUSTER_BACKEND_DOCKER}"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") cli_cluster_backend_arg"
    echo ""
    echo "  cli_cluster_backend_arg   Optional argument describing the cluster backend tests are running against. Must be one of 'docker', 'minikube' (default: ${DEFAULT_TESTSUITE_CLUSTER_BACKEND})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

testsuite_cluster_backend_arg="${1:-${DEFAULT_TESTSUITE_CLUSTER_BACKEND}}"
if [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ] &&
   [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    echo "Error: unknown cluster provided to run tests against. Must be one of 'docker', 'minikube'"
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

cd "${lang_root_dirpath}"
go build ./...

# The count=1 disables caching for the testsuite (which we want, because the engine server might have changed even though the test code didn't)
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  CGO_ENABLED=0 go test ./... -p "${PARALLELISM}" -count=1 -timeout "${TIMEOUT}" -tags minikube
else
  CGO_ENABLED=0 go test ./... -p "${PARALLELISM}" -count=1 -timeout "${TIMEOUT}"
fi

