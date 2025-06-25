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
	"github.com/datazip/olake-frontend/server/utils"
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
	platform      utils.PlatformInfo
	ipAddress     string
	locationInfo  *LocationInfo
	locationMutex sync.Mutex
	locationChan  chan struct{}
	anonymousID   string
	username      string
	wg            sync.WaitGroup
}

type LocationInfo struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

func loadTelemetryConfig() error {
	telemetryEnabled = constants.TelemetryEnabled
	segmentAPIKey = constants.TelemetrySegmentAPIKey

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

	// Generate anonymous ID during initialization
	anonymousID := generateStoredAnonymousID()

	instance = &Telemetry{
		client:       client,
		enabled:      telemetryEnabled,
		platform:     getPlatformInfo(),
		ipAddress:    ip,
		locationChan: make(chan struct{}),
		anonymousID:  anonymousID,
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

// generateStoredAnonymousID generates or retrieves a stored anonymous ID
func generateStoredAnonymousID() string {
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

func getPlatformInfo() utils.PlatformInfo {
	return utils.PlatformInfo{
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
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://ipinfo.io/%s/json", ip), http.NoBody)
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

// SetUsername sets the username for telemetry tracking
func SetUsername(username string) {
	if instance != nil {
		instance.username = username
	}
}

// TrackEvent sends a custom event to Segment
func TrackEvent(_ context.Context, eventName string, properties map[string]interface{}) error {
	fmt.Println("track event started", eventName)
	if instance == nil || !instance.enabled {
		return nil
	}

	if properties == nil {
		properties = make(map[string]interface{})
	}

	// Add username to properties if available
	if instance.username != "" {
		properties["username"] = instance.username
	}

	// Add common properties that needs to be sent with every event
	essentialProps := utils.GetTelemetryCommonProperties(instance.platform, instance.ipAddress, instance.getLocationWithTimeout())
	for key, value := range essentialProps {
		properties[key] = value
	}

	instance.wg.Add(1)
	go func() {
		defer func() {
			instance.client.Close()
		}()
		if err := instance.client.Enqueue(analytics.Track{
			UserId:     instance.anonymousID,
			Event:      eventName,
			Properties: properties,
		}); err != nil {
			// Log error but don't return it since we're in a goroutine
			fmt.Printf("Failed to send telemetry event %s: %v\n", eventName, err)
		}
	}()

	return nil
}
