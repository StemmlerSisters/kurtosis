package exec

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestExecInstruction_StringRepresentationWorks(t *testing.T) {
	execInstruction := NewExecInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		"example-service-id",
		[]string{"mkdir", "-p", "/tmp/store"},
		0,
	)
	expectedStr := `exec(command=["mkdir", "-p", "/tmp/store"], expected_exit_code=0, service_id="example-service-id")`
	require.Equal(t, expectedStr, execInstruction.String())
}
