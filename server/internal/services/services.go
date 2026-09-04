package services

import (
	"context"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/appconfig"
	"github.com/datazip-inc/olake-ui/server/internal/database"
	"github.com/datazip-inc/olake-ui/server/internal/services/etl"
	"github.com/datazip-inc/olake-ui/server/internal/services/optimization"
	"github.com/datazip-inc/olake-ui/server/internal/utils/logger"
	"github.com/datazip-inc/olake-ui/server/internal/utils/telemetry"
)

// how long the install id push to Fusion may take before it is given up on
const installIDPushTimeout = 30 * time.Second

type AppService struct {
	db  *database.Database
	etl *etl.Service
	opt *optimization.Service
}

func InitAppService(db *database.Database) (*AppService, error) {
	// Initialize ETL service
	etlSvc, err := etl.InitService(db)
	if err != nil {
		return nil, err
	}

	appSvc := &AppService{
		db:  db,
		etl: etlSvc,
		opt: nil,
	}

	enableOptimization := appconfig.Load().EnableOptimization
	if enableOptimization {
		optSvc, err := optimization.InitService()
		if err != nil {
			return nil, err
		}

		appSvc.opt = optSvc
		pushInstallIDToOptimization(optSvc)
	}

	return appSvc, nil
}

// pushInstallIDToOptimization tells Fusion which install id to report telemetry under. It runs in
// the background and its failures are only logged: telemetry must never delay or break startup,
// and the next UI start pushes the same id again.
func pushInstallIDToOptimization(opt *optimization.Service) {
	installID := telemetry.EnsureUserID()
	if installID == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), installIDPushTimeout)
		defer cancel()

		if err := opt.SetInstallID(ctx, installID); err != nil {
			logger.Warnf("failed to push telemetry install id to optimization: %s", err)
			return
		}
		logger.Info("telemetry install id shared with optimization")
	}()
}

func (s *AppService) ETL() *etl.Service {
	return s.etl
}

func (s *AppService) Optimization() *optimization.Service {
	return s.opt
}
