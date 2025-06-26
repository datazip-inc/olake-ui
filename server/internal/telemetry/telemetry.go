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
	client   analytics.Client
	idLock   sync.Mutex
	instance *Telemetry
)

type Telemetry struct {
	client       analytics.Client
	platform     utils.PlatformInfo
	ipAddress    string
	locationInfo *LocationInfo
	anonymousID  string
	username     string
}

type LocationInfo struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

func InitTelemetry() error {
	ip := getOutboundIP()
	client = analytics.New(constants.TelemetrySegmentAPIKey)
	// Generate anonymous ID during initialization
	anonymousID := generateStoredAnonymousID()

	instance = &Telemetry{
		client:      client,
		platform:    getPlatformInfo(),
		ipAddress:   ip,
		anonymousID: anonymousID,
	}

	if ip != constants.TelemetryIPNotFoundPlaceholder {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TelemetryConfigTimeout)
		defer cancel()
		loc, err := getLocationFromIP(ctx, ip)
		if err == nil {
			instance.locationInfo = &loc
		} else {
			fmt.Printf("Failed to fetch location for IP %s: %v\n", ip, err)
			instance.locationInfo = &LocationInfo{
				Country: "NA",
				Region:  "NA",
				City:    "NA",
			}
		}
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

// TrackEvent sends a custom event to Segment
func TrackEvent(_ context.Context, eventName string, properties map[string]interface{}) error {
	if instance == nil {
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
	essentialProps := utils.GetTelemetryCommonProperties(instance.platform, instance.ipAddress, instance.locationInfo)
	for key, value := range essentialProps {
		properties[key] = value
	}

	if err := instance.client.Enqueue(analytics.Track{
		UserId:     instance.anonymousID,
		Event:      eventName,
		Properties: properties,
	}); err != nil {
		fmt.Printf("Failed to send telemetry event %s: %v\n", eventName, err)
	}

	return nil
}

func Flush() {
	if instance != nil {
		time.Sleep(5 * time.Second)
		if err := instance.client.Close(); err != nil {
			fmt.Printf("Warning: Failed to close telemetry client: %v\n", err)
		}
	}
}

// GetStoredAnonymousID returns the stored anonymous ID from the config directory
func GetStoredAnonymousID() string {
	configDir := filepath.Join(os.TempDir(), "olake")
	idPath := filepath.Join(configDir, constants.TelemetryAnonymousIDFile)

	if idBytes, err := os.ReadFile(idPath); err == nil {
		return string(idBytes)
	}
	return ""
}

// SetUsername sets the username for telemetry tracking
func SetUsername(username string) {
	if instance != nil {
		instance.username = username
	}
}
