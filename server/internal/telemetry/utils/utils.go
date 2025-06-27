package utils

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OlakeVersion string
	OS           string
	Arch         string
	DeviceCPU    string
}

// GetTelemetryCommonProperties returns a map of common telemetry properties
func GetTelemetryCommonProperties(platform PlatformInfo, ipAddress string, location interface{}) map[string]interface{} {
	return map[string]interface{}{
		"olake_version": platform.OlakeVersion,
		"os":            platform.OS,
		"arch":          platform.Arch,
		"device_cpu":    platform.DeviceCPU,
		"ip_address":    ipAddress,
		"location":      location,
	}
}
