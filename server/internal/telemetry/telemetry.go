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
	"strconv"
	"time"

	analytics "github.com/segmentio/analytics-go/v3"
)

var instance *Telemetry

type LocationInfo struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

type PlatformInfo struct {
	OlakeVersion string
	OS           string
	Arch         string
	DeviceCPU    string
}

type Telemetry struct {
	client       analytics.Client
	platform     PlatformInfo
	ipAddress    string
	locationInfo *LocationInfo
	TempUserID   string
	username     string
}

func InitTelemetry() {
	if disabled, _ := strconv.ParseBool(os.Getenv("TELEMETRY_DISABLED")); disabled {
		return
	}

	go func() {
		if disabled, _ := strconv.ParseBool(os.Getenv("TELEMETRY_DISABLED")); disabled {
			return
		}

		ip := getOutboundIP()
		client := analytics.New(TelemetrySegmentAPIKey)

		// Generate user ID during initialization
		tempUserID := func() string {
			configDir := filepath.Join(os.TempDir(), "olake-config", "telemetry")
			idPath := filepath.Join(configDir, TelemetryUserIDFile)

			idBytes, err := os.ReadFile(idPath)

			if err != nil {
				newID := func() string {
					hash := sha256.New()
					hash.Write([]byte(time.Now().String()))
					return hex.EncodeToString(hash.Sum(nil))[:32]
				}()
				if err := os.MkdirAll(configDir, 0755); err != nil {
					return newID
				}
				_ = os.WriteFile(idPath, []byte(newID), 0600)
				return newID
			}
			return string(idBytes)
		}()

		instance = &Telemetry{
			client: client,
			platform: PlatformInfo{
				OS:           runtime.GOOS,
				Arch:         runtime.GOARCH,
				OlakeVersion: OlakeVersion,
				DeviceCPU:    fmt.Sprintf("%d cores", runtime.NumCPU()),
			},
			ipAddress:    ip,
			TempUserID:   tempUserID,
			locationInfo: getLocationFromIP(ip),
		}
	}()
}

func getOutboundIP() string {
	resp, err := http.Get(IPUrl)
	if err != nil {
		return IPNotFound
	}
	defer resp.Body.Close()

	ipBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return IPNotFound
	}

	return string(ipBody)
}

func getLocationFromIP(ip string) *LocationInfo {
	locationInfo := &LocationInfo{
		Country: "NA",
		Region:  "NA",
		City:    "NA",
	}

	if ip == IPNotFound || ip == "" {
		return locationInfo
	}

	ctx, cancel := context.WithTimeout(context.Background(), TelemetryConfigTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://ipinfo.io/%s/json", ip), http.NoBody)
	if err != nil {
		return locationInfo
	}

	client := http.Client{Timeout: TelemetryConfigTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return locationInfo
	}
	defer resp.Body.Close()

	var info struct {
		Country string `json:"country"`
		Region  string `json:"region"`
		City    string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return locationInfo
	}

	return &LocationInfo{
		Country: info.Country,
		Region:  info.Region,
		City:    info.City,
	}
}

// TrackEvent sends a custom event to Segment
func TrackEvent(_ context.Context, eventName string, properties map[string]interface{}) error {
	if properties == nil {
		properties = make(map[string]interface{})
	}

	// Add username to properties if available
	if instance.username != "" {
		properties["username"] = instance.username
	}

	props := map[string]interface{}{
		"olake_version": instance.platform.OlakeVersion,
		"os":            instance.platform.OS,
		"arch":          instance.platform.Arch,
		"device_cpu":    instance.platform.DeviceCPU,
		"ip_address":    instance.ipAddress,
		"location":      instance.locationInfo,
	}

	// Add common properties that needs to be sent with every event
	for key, value := range props {
		properties[key] = value
	}

	return instance.client.Enqueue(analytics.Track{
		UserId:     instance.TempUserID,
		Event:      eventName,
		Properties: properties,
	})
}

func Close() {
	if instance == nil {
		return
	}
	_ = instance.client.Close()
}

// SetUsername sets the username for telemetry tracking
func SetUsername(username string) {
	if instance != nil {
		instance.username = username
	}
}

func GetTelemetryUserID() string {
	if instance != nil {
		return instance.TempUserID
	}
	return ""
}
