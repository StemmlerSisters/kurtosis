module github.com/kurtosis-tech/container-engine-lib

go 1.15

require (
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200514193020-5da88705cccc+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/kurtosis-tech/object-attributes-schema-lib v0.0.0-20220215113756-abe2c0a2dd75
	github.com/kurtosis-tech/stacktrace v0.0.0-20211028211901-1c67a77b5409
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/protobuf v1.27.1 // indirect
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)
