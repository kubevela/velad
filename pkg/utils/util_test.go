package utils

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHaveController(t *testing.T) {
	testCases := []struct {
		Controllers string
		Controller  string
		isContained bool
	}{
		{
			Controllers: "*",
			Controller:  "deployment",
			isContained: true,
		},
		{
			Controllers: "*,-deployment",
			Controller:  "deployment",
			isContained: false,
		},
		{
			Controllers: "deployment",
			Controller:  "job",
			isContained: false,
		},
		{
			Controllers: "*,-job",
			Controller:  "job",
			isContained: false,
		},
	}
	for _, tc := range testCases {
		res := HaveController(tc.Controllers, tc.Controller)
		if res != tc.isContained {
			t.Fail()
		}
	}
}

func TestIsDeployByPod(t *testing.T) {
	testCases := []struct {
		Controllers   string
		isDeployByPod bool
	}{
		{Controllers: "*,-deployment", isDeployByPod: true},
		{Controllers: "*,-job", isDeployByPod: true},
		{Controllers: "deployment,job", isDeployByPod: true},
		{Controllers: "deployment,job,replicaset", isDeployByPod: false},
		{Controllers: "*", isDeployByPod: false},
	}
	for _, tc := range testCases {
		if IfDeployByPod(tc.Controllers) != tc.isDeployByPod {
			t.Fatal(tc.Controllers)
		}
	}

}

func TestVeladWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	w := VeladWriter{buf}
	_, err := w.Write([]byte(fmt.Sprintln("test")))
	assert.NoError(t, err)
	got := buf.String()
	assert.Equal(t, "test\n", got)
	_, err = w.Write([]byte(fmt.Sprintln("If you want to enable dashboard, please run \"vela addon enable velaux\"")))
	assert.NoError(t, err)
	got = buf.String()
	assert.Equal(t, fmt.Sprintf("test\nIf you want to enable dashboard, please run \"vela addon enable %s\"\n", velauxDir), got)
}

func TestGetTmpDir(t *testing.T) {
	tmpDir,err := GetTmpDir()
	assert.NoError(t, err)
	fmt.Println(tmpDir)
	assert.NotEmpty(t, tmpDir)
}