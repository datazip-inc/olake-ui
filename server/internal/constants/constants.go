package constants

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/server/web"
	"github.com/spf13/viper"
)

var (
	defaultPort     = 8080
	version         = "Not Set"
	commitsha       = "Not Set"
	releasechannel  = "Not Set"
	defaultBaseHost = "localhost"
	DefaultTimeZone = "Asia/Kolkata"
	DefaultUsername = "olake"
	DefaultPassword = "password"
	TableNameMap    = map[TableType]string{}
)

// Telemetry event constants
const (
	EventUserLogin           = "user_login"
	EventJobCreated          = "job_created"
	EventSyncStarted         = "sync_started"
	EventSyncCompleted       = "sync_completed"
	EventSyncFailed          = "sync_failed"
	EventSourceCreated       = "source_created"
	EventDestinationCreated  = "destination_created"
	EventSourcesUpdated      = "sources_updated"
	EventDestinationsUpdated = "destinations_updated"
)

// Telemetry configuration constants
const (
	TelemetryAnonymousIDFile       = "telemetry_id"
	TelemetryVersion               = "0.0.1"
	TelemetryIPNotFoundPlaceholder = "NA"
	TelemetryConfigTimeout         = time.Second
	TelemetryLocationTimeout       = 200 * time.Millisecond
	TelemetryEnabled               = true
	TelemetrySegmentAPIKey         = "1gZZyBlRTkwWnyJPanBYnQ5E4cQwS6T6"
)

var RequiredConfigVariable = []string{"postgresdb", "copyrequestbody", "logsdir"}

func Init() {
	viper.AutomaticEnv()

	viper.SetDefault("PORT", defaultPort)
	viper.SetDefault("BUILD", version)
	viper.SetDefault("COMMITSHA", commitsha)
	viper.SetDefault("RELEASE_CHANNEL", releasechannel)
	viper.SetDefault("BASE_HOST", defaultBaseHost)
	viper.SetDefault("BASE_URL", fmt.Sprintf("%s:%v", viper.GetString("BASE_HOST"), viper.GetString("PORT")))

	checkForRequiredVariables(RequiredConfigVariable)

	// init table names
	TableNameMap = map[TableType]string{
		UserTable:        "olake-$$-user",
		SourceTable:      "olake-$$-source",
		DestinationTable: "olake-$$-destination",
		JobTable:         "olake-$$-job",
		CatalogTable:     "olake-$$-catalog",
		SessionTable:     "session",
	}

	// replace $$ with the environment
	for k, v := range TableNameMap {
		TableNameMap[k] = strings.ReplaceAll(v, "$$", web.BConfig.RunMode)
	}
}

func checkForRequiredVariables(vars []string) {
	for _, v := range vars {
		value, err := config.String(v)
		if err != nil || value == "" {
			panic("Required config variable not found: ," + v)
		}
	}
}
