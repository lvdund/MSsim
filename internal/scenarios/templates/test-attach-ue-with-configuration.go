package templates

import "mssim/config"

func TestAttachUeWithConfiguration(tunnelEnabled bool) {
	tunnelMode := config.TunnelDisabled
	if tunnelEnabled {
		tunnelMode = config.TunnelVrf
	}
	TestMultiUesInQueue(1, tunnelMode, true, false, 500, 0, 0, 0, 0, 0, 1, "")
}
