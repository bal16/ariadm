//go:build !windows

package daemon

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStop_TerminatesProcessGracefully(t *testing.T) {
	dm := &DaemonManager{cmd: exec.Command("sleep", "30")}
	require.NoError(t, dm.cmd.Start())

	started := time.Now()
	require.NoError(t, dm.Stop())

	assert.Nil(t, dm.cmd)
	assert.Less(t, time.Since(started), 4*time.Second, "Stop must not wait for the full shutdown timeout")
}
