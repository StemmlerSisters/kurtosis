module github.com/kurtosis-tech/kurtosis-engine-server/server

go 1.15

replace (
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang => ../api/golang
	github.com/kurtosis-tech/kurtosis-engine-server/launcher => ../launcher
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/container-engine-lib v0.0.0-20211116225347-a5bd1c49b423
	github.com/kurtosis-tech/free-ip-addr-tracker-lib v0.0.0-20211106222342-d3be9e82993e
	github.com/kurtosis-tech/kurtosis-core/launcher v0.0.0-20211216182312-4b8a5841426e
	github.com/kurtosis-tech/kurtosis-engine-server/api/golang v0.0.0
	github.com/kurtosis-tech/kurtosis-engine-server/launcher v0.0.0
	github.com/kurtosis-tech/metrics-library/golang v0.0.0-20220121190509-e4bc02b1263c // indirect
	github.com/kurtosis-tech/minimal-grpc-server/golang v0.0.0-20211201000847-a204edc5a0b3
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20211206192525-530b79739d16
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)
