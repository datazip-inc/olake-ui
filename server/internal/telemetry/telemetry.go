package telemetry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/datazip/olake-frontend/server/internal/constants"
	analytics "github.com/segmentio/analytics-go/v3"
)



var (
	client           analytics.Client
	idLock           sync.Mutex
	telemetryEnabled bool
	segmentAPIKey    string
	instance         *Telemetry
)

type Telemetry struct {
	client        analytics.Client
	enabled       bool
	platform      platformInfo
	ipAddress     string
	locationInfo  *LocationInfo
	locationMutex sync.Mutex
	locationChan  chan struct{}
}

type platformInfo struct {
	OS           string
	Arch         string
	Environment  string
	OlakeVersion string
	DeviceCPU    string
}

type LocationInfo struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

func loadTelemetryConfig() error {
	// Hardcoded telemetry configuration
	telemetryEnabled = true
	segmentAPIKey = "1gZZyBlRTkwWnyJPanBYnQ5E4cQwS6T6"
	
	if telemetryEnabled && segmentAPIKey == "" {
		return fmt.Errorf("segment API key is required when telemetry is enabled")
	}
	
	return nil
}

func InitTelemetry() error {
	if err := loadTelemetryConfig(); err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	ip := getOutboundIP()

	if telemetryEnabled {
		client = analytics.New(segmentAPIKey)
	}

	instance = &Telemetry{
		client:       client,
		enabled:      telemetryEnabled,
		platform:     getPlatformInfo(),
		ipAddress:    ip,
		locationChan: make(chan struct{}),
	}

	if instance.enabled && ip != constants.TelemetryIPNotFoundPlaceholder {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), constants.TelemetryConfigTimeout)
			defer cancel()
			location, err := getLocationFromIP(ctx, ip)
			if err == nil {
				instance.locationMutex.Lock()
				instance.locationInfo = &location
				instance.locationMutex.Unlock()
			}
			close(instance.locationChan)
		}()
	} else {
		close(instance.locationChan)
	}

	return nil
}

func GetAnonymousID() string {
	idLock.Lock()
	defer idLock.Unlock()

	configDir := getConfigDir()
	idPath := filepath.Join(configDir, constants.TelemetryAnonymousIDFile)

	if idBytes, err := os.ReadFile(idPath); err == nil {
		return string(idBytes)
	}

	newID := generateUUID()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return newID
	}
	_ = os.WriteFile(idPath, []byte(newID), 0600)
	return newID
}

func getConfigDir() string {
	return filepath.Join(os.TempDir(), "olake")
}

func generateUUID() string {
	hash := sha256.New()
	hash.Write([]byte(time.Now().String()))
	return hex.EncodeToString(hash.Sum(nil))[:32]
}

func getPlatformInfo() platformInfo {
	return platformInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		OlakeVersion: constants.TelemetryVersion,
		DeviceCPU:    fmt.Sprintf("%d cores", runtime.NumCPU()),
	}
}

func getOutboundIP() string {
	ip := []byte(constants.TelemetryIPNotFoundPlaceholder)
	resp, err := http.Get("https://api.ipify.org?format=text")

	if err != nil {
		return string(ip)
	}

	defer resp.Body.Close()
	ipBody, err := io.ReadAll(resp.Body)
	if err == nil {
		ip = ipBody
	}

	return string(ip)
}

func getLocationFromIP(ctx context.Context, ip string) (LocationInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://ipinfo.io/%s/json", ip), nil)
	if err != nil {
		return LocationInfo{}, err
	}

	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return LocationInfo{}, err
	}
	defer resp.Body.Close()

	var info struct {
		Country string `json:"country"`
		Region  string `json:"region"`
		City    string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return LocationInfo{}, err
	}

	return LocationInfo{
		Country: info.Country,
		Region:  info.Region,
		City:    info.City,
	}, nil
}

func (t *Telemetry) getLocationWithTimeout() interface{} {
	select {
	case <-t.locationChan: // Returns immediately if channel already closed
	case <-time.After(constants.TelemetryLocationTimeout):
	}

	t.locationMutex.Lock()
	defer t.locationMutex.Unlock()

	if t.locationInfo != nil {
		return t.locationInfo
	}
	return constants.TelemetryIPNotFoundPlaceholder
}

// TrackEvent sends a custom event to Segment
func TrackEvent(ctx context.Context, eventName string, properties map[string]interface{}) error {
	if instance == nil || !instance.enabled {
		return nil
	}

	if properties == nil {	
		properties = make(map[string]interface{})
	}

	// Add essential properties
	properties["olake_version"] = instance.platform.OlakeVersion
	properties["os"] = instance.platform.OS
	properties["arch"] = instance.platform.Arch
	properties["device_cpu"] = instance.platform.DeviceCPU
	properties["ip_address"] = instance.ipAddress
	properties["location"] = instance.getLocationWithTimeout()
	properties["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	return instance.client.Enqueue(analytics.Track{
		UserId:     GetAnonymousID(),
		Event:      eventName,
		Properties: properties,
	})
}

// Flush ensures all events are sent before shutdown
func Flush() {
	if instance != nil && instance.client != nil {
		instance.client.Close()
	}
} 