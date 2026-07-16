package connections

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	helperconfig "github.com/radian-solusi/go-helpers/config"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const dateLayout = "2006-01-02"

type gormWrapper struct {
	db *gorm.DB
}

// getGormLogger returns a GORM logger based on GIN_MODE.
func newGORMLogger() logger.Interface {
	var logWriter io.Writer
	var logLevel logger.LogLevel

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		logsDir := "logs"
		os.MkdirAll(logsDir, 0755)
		logFileName := filepath.Join(logsDir, "database-"+time.Now().Format(dateLayout)+".log")
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open database log file, using stdout: %v", err)
			logWriter = os.Stdout
		} else {
			logWriter = logFile
		}
		logLevel = logger.Warn
	} else {
		logWriter = os.Stdout
		logLevel = logger.Info
	}

	return logger.New(
		log.New(logWriter, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(200) * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  ginMode != "release",
		},
	)
}

func NewDatabase(cfg helperconfig.DatabaseConfig) (Database, error) {
	var dialector gorm.Dialector
	switch cfg.Type {
	case helperconfig.DBTypePostgres:
		dialector = postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port))
	case helperconfig.DBTypeMySQL:
		dialector = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName))
	case helperconfig.DBTypeSQLite:
		if cfg.DBName == "" {
			return nil, errors.New("database name is required")
		}
		if dir := filepath.Dir(cfg.DBName); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create database directory: %w", err)
			}
		}
		dialector = sqlite.Open(cfg.DBName + "?cache=shared&mode=rwc")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 newGORMLogger(),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", cfg.Type, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get SQL database: %w", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, fmt.Errorf("register otelgorm: %w", err)
	}
	return &gormWrapper{db: db}, nil
}

func (g *gormWrapper) DB() *gorm.DB { return g.db }

func (g *gormWrapper) Ping(ctx context.Context) error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return fmt.Errorf("get SQL database: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

func (g *gormWrapper) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return fmt.Errorf("get SQL database: %w", err)
	}
	return sqlDB.Close()
}

func (g *gormWrapper) Migrate(models ...any) error {
	if len(models) == 0 {
		return errors.New("migration models are required")
	}
	if err := g.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}

func (g *gormWrapper) SetProgressMode(enabled bool) {
	if enabled {
		g.db.Logger = logger.New(
			log.New(io.Discard, "", 0),
			logger.Config{
				SlowThreshold:             0,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	} else {
		g.db.Logger = newGORMLogger()
	}
}
