package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db/migrate"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var currentDBType string
var currentDSN string

func InitDB(dbType, dsn string, debug bool) error {
	var err error
	currentDBType = dbType
	currentDSN = dsn
	gormConfig := gorm.Config{Logger: logger.Discard}
	if debug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	switch dbType {
	case "sqlite":
		db, err = initSQLite(dsn, &gormConfig)
	case "mysql":
		db, err = initMySQL(dsn, &gormConfig)
	case "postgres", "postgresql":
		db, err = initPostgres(dsn, &gormConfig)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := migrate.BeforeAutoMigrate(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Channel{},
		&model.ChannelKey{},
		&model.Group{},
		&model.GroupItem{},
		&model.LLMInfo{},
		&model.RouteTargetOverride{},
		&model.UpstreamSite{},
		&model.UpstreamCredential{},
		&model.UpstreamKeySnapshot{},
		&model.UpstreamGroupSnapshot{},
		&model.UpstreamModelPrice{},
		&model.AITask{},
		&model.AITaskStep{},
		&model.AIPromptTemplate{},
		&model.AIProfile{},
		&model.AIProfileVersion{},
		&model.AIGroupingProfile{},
		&model.AIChannelRecognitionProfile{},
		&model.AIPriceRecognitionProfile{},
		&model.AIModelClassificationProfile{},
		&model.AIConfigHealthProfile{},
		&model.DynamicRouteLearningState{},
		&model.GovernanceSession{},
		&model.GovernanceApplyRun{},
		&model.GovernanceRollbackPoint{},
		&model.StrategyProfile{},
		&model.APIKey{},
		&model.Setting{},
		&model.StatsTotal{},
		&model.StatsDaily{},
		&model.StatsHourly{},
		&model.StatsModel{},
		&model.StatsChannel{},
		&model.StatsAPIKey{},
		&model.OpsMetricBucket{},
		&model.RelayLog{},
		&migrate.MigrationRecord{},
	); err != nil {
		return err
	}
	if err := migrate.AfterAutoMigrate(db); err != nil {
		return err
	}
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		if err := db.Exec("DEALLOCATE ALL").Error; err != nil {
			log.Warnf("postgres DEALLOCATE ALL failed: %v", err)
		}
		if err := db.Exec("DISCARD ALL").Error; err != nil {
			log.Warnf("postgres DISCARD ALL failed: %v", err)
		}
	}
	return nil
}

func initSQLite(path string, config *gorm.Config) (*gorm.DB, error) {
	params := []string{
		"_journal_mode=WAL",
		"_synchronous=NORMAL",
		"_cache_size=10000",
		"_busy_timeout=5000",
		"_foreign_keys=ON",
		"_auto_vacuum=INCREMENTAL",
		"_mmap_size=268435456",
		"_locking_mode=NORMAL",
	}
	return gorm.Open(sqlite.Open(path+"?"+strings.Join(params, "&")), config)
}

func initMySQL(dsn string, config *gorm.Config) (*gorm.DB, error) {
	if !strings.Contains(dsn, "?") {
		dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
	}
	return gorm.Open(mysql.Open(dsn), config)
}

func initPostgres(dsn string, config *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), config)
}

func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func GetDB() *gorm.DB {
	return db
}

func GetCurrentDBType() string {
	return currentDBType
}

func GetCurrentDSN() string {
	return currentDSN
}
