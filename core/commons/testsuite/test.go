package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"time"
)

/*
An interface encapsulating a test to run against a test network.
 */
type Test interface {
	// NOTE: if Go had generics, 'network' would be a parameterized type representing the network that this test consumes
	// as produced by the NetworkLoader
	/*
	Runs test logic against the given network, with failures reported using the given context.
	 */
	Run(network networks.Network, context TestContext)

	// Gets the network loader that will be used to spin up the test network that the test will run against
	GetNetworkLoader() (networks.NetworkLoader, error)

	// The amount of time the test Run method will be allowed to execute for before it's killed and the test marked as failed
	// This does NOT include the time needed to do pre-test setup or post-test teardown, which is handled by GetSetupBuffer
	GetExecutionTimeout() time.Duration

	// How long the test will be given to do the pre-execution setup and post-setup teardown before the test will be hard-killed
	// The total amount of time a given test is allowed to run for *on top of* the GetExecutionTimeout value
	GetSetupBuffer() time.Duration
}
