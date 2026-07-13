package cmd

import (
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/observability"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/handlers"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/shutdown"
	"github.com/spf13/cobra"
)

var cfgFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start " + conf.APP_NAME,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		conf.PrintBanner()
		if err := conf.Load(cfgFile); err != nil {
			return err
		}
		log.SetLevel(conf.AppConfig.Log.Level)
		log.Configure(conf.AppConfig.Log.File, conf.AppConfig.Log.MaxSize, conf.AppConfig.Log.MaxBackups, conf.AppConfig.Log.MaxAge)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		shutdown.Init(log.Logger)
		if err := db.InitDB(conf.AppConfig.Database.Type, conf.AppConfig.Database.Path, conf.IsDebug()); err != nil {
			log.Errorf("database init error: %v", err)
			return
		}
		shutdown.Register(db.Close)

		handlers.SetAuditRecorder(observability.NewAuditRecorder(db.GetDB()))

		if err := op.InitCache(); err != nil {
			log.Errorf("cache init error: %v", err)
			shutdown.Shutdown()
			return
		}
		shutdown.Register(op.CancelAllAITasks)
		shutdown.Register(op.SaveCache)

		if err := server.Start(); err != nil {
			log.Errorf("server start error: %v", err)
			shutdown.Shutdown()
			return
		}
		shutdown.Register(server.Close)

		task.Init()
	shutdown.Register(task.StopAll)
	op.StartHealthCheckTask()
	op.StartResponseCacheEvictTask()
	op.StartBackupTask()
	go task.RUN()

		shutdown.Listen()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
