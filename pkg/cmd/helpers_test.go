package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oam-dev/velad/pkg/apis"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	f()
	assert.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	assert.NoError(t, err)
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
