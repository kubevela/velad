package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oam-dev/velad/pkg/apis"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = r.Close() }()
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	f()
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func TestPrintClusterStatusK3sShowsReason(t *testing.T) {
	status := apis.ClusterStatus{
		K3s: apis.K3sStatus{
			K3sBinary:        true,
			K3sServiceStatus: "active",
			Reason:           "fail to run systemctl: not found",
		},
	}
	out := captureStdout(t, func() {
		printClusterStatusK3s(status)
	})
	assert.Contains(t, out, "fail to run systemctl: not found")
}
