//go:build linux

package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oam-dev/velad/pkg/apis"
)

func TestFillServiceStatusSystemctlNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	status := apis.ClusterStatus{}
	fillServiceStatus(&status)
	assert.Contains(t, status.K3s.Reason, "executable file not found")
	assert.NotContains(t, status.K3s.Reason, "<nil>")
}
