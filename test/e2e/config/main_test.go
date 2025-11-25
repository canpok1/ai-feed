//go:build e2e

package config

import (
	"os"
	"testing"

	"github.com/canpok1/ai-feed/test/e2e/common"
)

func TestMain(m *testing.M) {
	common.SetupPackage()
	code := m.Run()
	common.CleanupPackage()
	os.Exit(code)
}
