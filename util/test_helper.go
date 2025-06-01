package util

import (
	"os"
	"testing"
)

type TestHelper struct{}

func (helper *TestHelper) ReadFile(t *testing.T, path string) string {
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Read error: %+v\n", err)
		return ""
	}

	return string(buf)
}
