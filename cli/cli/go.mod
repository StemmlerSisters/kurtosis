module github.com/kurtosis-tech/kurtosis-cli/cli

go 1.15

replace github.com/kurtosis-tech/kurtosis-cli/commons => ../commons

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/adrg/xdg v0.4.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gammazero/workerpool v1.1.2
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20220222221450-857e15a05dbc
	github.com/kurtosis-tech/kurtosis-cli/commons v0.0.0 // Local dependency
	github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang v0.0.0-20220222221856-130cdb57cec6
	github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang v0.0.0-20220222223007-4482283840f4
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0-20220222222849-bc5435915254
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20220215151652-4f1a58645739
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220215113756-abe2c0a2dd75
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)
