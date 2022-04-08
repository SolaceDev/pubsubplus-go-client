//go:build dummy
// +build dummy

// workaround for vendoring
package lib

import (
	_ "solace.dev/go/messaging/internal/ccsmp/lib/darwin"
	_ "solace.dev/go/messaging/internal/ccsmp/lib/include/solclient"
	_ "solace.dev/go/messaging/internal/ccsmp/lib/linux"
)
