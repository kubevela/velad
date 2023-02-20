package utils

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	tmpDir, err := GetTmpDir()
	assert.NoError(t, err)
	fmt.Println(tmpDir)
	assert.NotEmpty(t, tmpDir)
}
