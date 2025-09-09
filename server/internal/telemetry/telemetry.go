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
	"strings"
	"time"
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
	httpClient   *http.Client
	platform     PlatformInfo
	ipAddress    string
	locationInfo *LocationInfo
	TempUserID   string
	username     string
}

func InitTelemetry() {
	go func() {
		if disabled, _ := strconv.ParseBool(os.Getenv("TELEMETRY_DISABLED")); disabled {
			return
		}

		ip := getOutboundIP()

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
			httpClient: &http.Client{Timeout: TelemetryClientTimeout},
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
	if instance.httpClient == nil {
		return fmt.Errorf("telemetry client is nil")
	}

	if properties == nil {
		properties = make(map[string]interface{})
	}
	properties["olake_version"] = instance.platform.OlakeVersion
	properties["os"] = instance.platform.OS
	properties["arch"] = instance.platform.Arch
	properties["device_cpu"] = instance.platform.DeviceCPU
	properties["ip_address"] = instance.ipAddress
	properties["location"] = instance.locationInfo
	properties["distinct_id"] = instance.TempUserID
	properties["time"] = time.Now().Unix()
	properties["event_original_name"] = eventName

	// Add username to properties if available
	if instance.username != "" {
		properties["username"] = instance.username
	}

	body := map[string]interface{}{
		"event":      eventName,
		"properties": properties,
	}
	propsBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), TelemetryConfigTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", ProxyTrackURL, strings.NewReader(string(propsBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := instance.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send telemetry event, status: %s, response: %s", resp.Status, string(respBody))
	}
	return nil
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
