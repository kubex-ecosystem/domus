package cli

import (
	"context"

	"github.com/kubex-ecosystem/domus/internal/backends"
	"github.com/kubex-ecosystem/domus/internal/engine"
	kbxMod "github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/provider"
	systemservice "github.com/kubex-ecosystem/domus/internal/services/system_service"

	"github.com/kubex-ecosystem/domus/internal/datastore/extensions"
	_ "github.com/kubex-ecosystem/domus/internal/datastore/extensions/schemas"

	"github.com/kubex-ecosystem/ethyr/tools/flow"
	"github.com/kubex-ecosystem/ethyr/tools/flow/fsm"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

var logger = gl.GetLoggerZ("Migration")

// Estados Atômicos da Migração Domus
const (
	MigrateIdle fsm.State = 1 << iota
	MigrateConfigLoading
	MigrateBackendLoading
	MigrateExecuting
	MigrateDone
	MigrateError
)

// Eventos de Transição
const (
	EvStartMigrate fsm.Event = iota
	EvConfigReady
	EvBackendReady
	EvMigrationSuccess
	EvMigrationFail
)

func DatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}

	cmd := &cobra.Command{
		Use:   "database",
		Short: "Database management commands for Domus",
		Run: func(cmd *cobra.Command, args []string) {
			if initArgs.Debug {
				gl.SetDebugMode(initArgs.Debug)
			}
			gl.Info("Domus", "Database management commands for Domus")
			cmd.Help()
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	cmd.AddCommand(startDatabaseCmd())
	cmd.AddCommand(stopDatabaseCmd())
	cmd.AddCommand(statusDatabaseCmd())
	cmd.AddCommand(migrateDatabaseCmd())

	return cmd
}

func startDatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Database services",
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(initArgs.Debug)
			if err := migrateDatabaseCmd().Execute(); err != nil {
				return gl.Errorf("Error executing migration: %v", err)
			}
			if err := systemservice.StartSystemService(initArgs); err != nil {
				return gl.Errorf("Error starting database services: %v", err)
			}
			return nil
		},
	}
	return cmd
}

func stopDatabaseCmd() *cobra.Command {
	return &cobra.Command{Use: "stop", Run: func(cmd *cobra.Command, args []string) { cmd.Help() }}
}

func statusDatabaseCmd() *cobra.Command {
	return &cobra.Command{Use: "status", Run: func(cmd *cobra.Command, args []string) { cmd.Help() }}
}

func migrateDatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}
	var keepAlive bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			gl.SetDebugMode(initArgs.Debug)

			// Inicializa o Maestro Atômico para a Migração
			m := flow.NewFlow()
			m.FSM = fsm.NewFSM(MigrateIdle, []fsm.Transition{
				{From: MigrateIdle, Event: EvStartMigrate, To: MigrateConfigLoading},
				{From: MigrateConfigLoading, Event: EvConfigReady, To: MigrateBackendLoading},
				{From: MigrateBackendLoading, Event: EvBackendReady, To: MigrateExecuting},
				{From: MigrateExecuting, Event: EvMigrationSuccess, To: MigrateDone},
				{From: MigrateExecuting, Event: EvMigrationFail, To: MigrateError},
			})

			// Inicia o Processo
			if err := m.State.Start(); err != nil {
				return err
			}
			m.FSM.Trigger(EvStartMigrate)
			gl.Infof("[Atomic Migration: %d] Iniciando...", m.FSM.Current())

			// 1. LOAD CONFIG
			configPath := kbxGet.ValOrType(initArgs.ConfigFile, kbxGet.EnvOr("KUBEX_DOMUS_CONFIG_FILE", kbxMod.DefaultConfigFile))
			rootConfig, err := engine.LoadRootConfig(configPath)
			if err != nil {
				m.FSM.Trigger(EvMigrationFail)
				return gl.Errorf("Failed to load config: %v", err)
			}
			m.FSM.Trigger(EvConfigReady)
			gl.Infof("[Atomic Migration: %d] Configuração carregada.", m.FSM.Current())

			var dbConfig *kbxMod.DBConfig
			if len(initArgs.DBConfigID) > 0 {
				dbConfig = engine.GetDBConfig(&rootConfig, initArgs.DBConfigID)
			} else {
				dbConfig = engine.GetDefaultDBConfig(&rootConfig)
			}
			if dbConfig == nil {
				m.FSM.Trigger(EvMigrationFail)
				return gl.Errorf("No default database configuration found")
			}

			// 2. BACKEND STACK
			backendStack, ok := backends.GetProvider(dbConfig.Backend)
			if !ok {
				m.FSM.Trigger(EvMigrationFail)
				return gl.Errorf("Backend '%s' not found", dbConfig.Backend)
			}
			m.FSM.Trigger(EvBackendReady)
			gl.Infof("[Atomic Migration: %d] Backend selecionado.", m.FSM.Current())

			migratableStack, ok := backendStack.(provider.MigratableProvider)
			if !ok {
				m.FSM.Trigger(EvMigrationSuccess) // Not an error, just skip
				gl.Info("Backend does not support migrations. skipping.")
				return nil
			}

			// 3. EXECUTE MIGRATION
			endpoints, err := migratableStack.Start(ctx, provider.ConvertRootConfigToStartSpec(&rootConfig))
			if err != nil {
				m.FSM.Trigger(EvMigrationFail)
				return err
			}

			dbMgr := engine.NewDatabaseManager(logger)
			for _, endpoint := range endpoints {
				if endpoint.DBConfig.Migration == nil {
					continue
				}
				conn, err := dbMgr.LoadDBConfig(endpoint.DBConfig)
				if err != nil {
					m.FSM.Trigger(EvMigrationFail)
					return err
				}
				if err := migratableStack.PrepareMigrations(ctx, &conn); err != nil {
					m.FSM.Trigger(EvMigrationFail)
					return err
				}
				if err := migratableStack.RunMigrations(ctx, &conn, endpoint.DBConfig.Migration); err != nil {
					m.FSM.Trigger(EvMigrationFail)
					return err
				}

				// Apply all registered schema extensions
				exec, err := conn.Driver.Executor(ctx)
				if err == nil && exec != nil {
					if extErr := extensions.Global.ApplyAll(ctx, exec); extErr != nil {
						gl.Warnf("Failed to apply schema extensions: %v", extErr)
					}
				}
			}

			m.FSM.Trigger(EvMigrationSuccess)
			m.State.Complete()
			gl.Successf("[Atomic Migration: %d] Migração concluída com sucesso!", m.FSM.Current())

			return nil
		},
	}

	cmd.Flags().StringVarP(&initArgs.DBConfigID, "target-db", "t", "", "Target database config ID")
	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().BoolVarP(&keepAlive, "keep-alive", "k", false, "Keep engine connections alive")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "", "Path to configuration file")

	return cmd
}
