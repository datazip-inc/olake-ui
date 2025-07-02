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
	"time"

	"github.com/datazip/olake-frontend/server/internal/telemetry/utils"
	analytics "github.com/segmentio/analytics-go/v3"
)

var (
	client   analytics.Client
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
	client = analytics.New(utils.TelemetrySegmentAPIKey)
	// Generate anonymous ID during initialization
	anonymousID := generateStoredAnonymousID()

	instance = &Telemetry{
		client:      client,
		platform:    getPlatformInfo(),
		ipAddress:   ip,
		anonymousID: anonymousID,
	}

	if ip == utils.TelemetryIPNotFoundPlaceholder {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), utils.TelemetryConfigTimeout)
	defer cancel()
	loc, err := getLocationFromIP(ctx, ip)

	if err != nil {
		fmt.Printf("Failed to fetch location for IP %s: %s\n", ip, err)
		instance.locationInfo = &LocationInfo{
			Country: "NA",
			Region:  "NA",
			City:    "NA",
		}
	}
	instance.locationInfo = &loc
	return nil
}

// generateStoredAnonymousID generates or retrieves a stored anonymous ID
func generateStoredAnonymousID() string {
	configDir := getConfigDir()
	idPath := filepath.Join(configDir, utils.TelemetryAnonymousIDFile)

	idBytes, err := os.ReadFile(idPath)

	if err != nil {
		newID := generateUUID()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return newID
		}
		_ = os.WriteFile(idPath, []byte(newID), 0600)
		return newID
	}
	return string(idBytes)
}

func getConfigDir() string {
	return filepath.Join(os.TempDir(), "olake-config", "telemetry")
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
		OlakeVersion: utils.TelemetryVersion,
		DeviceCPU:    fmt.Sprintf("%d cores", runtime.NumCPU()),
	}
}

func getOutboundIP() string {
	resp, err := http.Get("https://api.ipify.org?format=text")

	if err != nil {
		return utils.TelemetryIPNotFoundPlaceholder
	}

	defer resp.Body.Close()

	ipBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.TelemetryIPNotFoundPlaceholder
	}

	return string(ipBody)
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

	err := instance.client.Enqueue(analytics.Track{
		UserId:     instance.anonymousID,
		Event:      eventName,
		Properties: properties,
	})

	if err != nil {
		fmt.Printf("Failed to send telemetry event %s: %s\n", eventName, err)
	}

	return nil
}

func Flush() {
	if instance == nil {
		return
	}
	time.Sleep(5 * time.Second)
	err := instance.client.Close()
	if err != nil {
		fmt.Printf("Warning: Failed to close telemetry client: %s\n", err)
	}
}

// SetUsername sets the username for telemetry tracking
func SetUsername(username string) {
	if instance != nil {
		instance.username = username
	}
}
