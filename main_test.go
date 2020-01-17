package agollo

import (
	"os"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	// setup

	code := m.Run()
	os.Exit(code)
}

var (
	bkOption    *option
	bkInitOnce  sync.Once
	bkStartOnce sync.Once
	bkService   *service
	bkDefault   map[string]defaultVal
	bkCallback  CHandler
)

func backupGlobals() {
	bkOption = gOption
	bkInitOnce = gInitOnce
	bkStartOnce = gStartOnce
	bkService = gService
	bkDefault = gDefault
	bkCallback = gCallback
}

func recoverGlobals() {
	gOption = bkOption
	gInitOnce = bkInitOnce
	gStartOnce = bkStartOnce
	gService = bkService
	gDefault = bkDefault
	gCallback = bkCallback
}
