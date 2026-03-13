package cli

import (
	"github.com/kubex-ecosystem/domus/internal/engine"
	"github.com/kubex-ecosystem/domus/internal/provider"

	"context"
	"os"

	"github.com/kubex-ecosystem/domus/internal/module/info"

	"github.com/spf13/cobra"

	"github.com/kubex-ecosystem/domus/internal/backends"

	kbxMod "github.com/kubex-ecosystem/domus/internal/module/kbx"
	systemservice "github.com/kubex-ecosystem/domus/internal/services/system_service"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxInfo "github.com/kubex-ecosystem/kbx/tools/info"
	gl "github.com/kubex-ecosystem/logz"
)

var logger = gl.GetLoggerZ("Migration")

func DatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}

	shortDesc := "Database management commands for Domus"
	longDesc := "Database management commands for Domus"
	cmd := &cobra.Command{
		Use:         "database",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("KUBEX_DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if initArgs.Debug {
				gl.SetDebugMode(initArgs.Debug)
			}
			gl.Info("Domus", "Database management commands for Domus")
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
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

	shortDesc := "Start Database services"
	longDesc := "Start Database services. This will launch the database in a Docker container, and keep a minimal Z"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("KUBEX_DOMUS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(initArgs.Debug)
			if err := migrateDatabaseCmd().Execute(); err != nil {
				return gl.Errorf("Error executing migration: %v", err)
			}

			if err := systemservice.StartSystemService(initArgs); err != nil {
				return gl.Errorf("Error starting database services: %v", err)
			}
			gl.Info("Domus", "Database services started successfully.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	return cmd
}

func stopDatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}

	shortDesc := "Stop Docker"
	longDesc := "Stop Docker service"

	cmd := &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("KUBEX_DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	return cmd
}

func statusDatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}

	shortDesc := "Status Docker"
	longDesc := "Status Docker service"

	cmd := &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("KUBEX_DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	return cmd
}

func migrateDatabaseCmd() *cobra.Command {
	var initArgs = &kbxMod.InitArgs{}
	var keepAlive bool

	shortDesc := "Run database migrations"
	longDesc := "Run database migrations for all registered models."

	cmd := &cobra.Command{
		Use:         "migrate",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("KUBEX_DOMUS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize context and logger
			ctx := context.Background()
			gl.SetDebugMode(initArgs.Debug)

			// ========== STEP 1: LOAD CONFIG ==========
			gl.Info("Loading configuration...")
			rootConfig, err := engine.LoadRootConfig(kbxGet.ValOrType(initArgs.ConfigFile, kbxGet.EnvOr("KUBEX_DOMUS_CONFIG_FILE", kbxMod.DefaultConfigFile)))
			if err != nil {
				return gl.Errorf("Failed to load config: %v", err)
			}

			var dbConfig *kbxMod.DBConfig
			if len(initArgs.DBConfigID) > 0 {
				dbConfig = engine.GetDBConfig(&rootConfig, initArgs.DBConfigID)
			} else {
				dbConfig = engine.GetDefaultDBConfig(&rootConfig)
			}
			if dbConfig == nil {
				return gl.Errorf("No default database configuration found")
			}

			// ========== STEP 2: CREATE BACKEND STACK ==========
			gl.Info("Loading backend stack...")

			providers := backends.ListProviders()
			for _, provider := range providers {
				gl.Debugf("Provider Registered: %s", provider.Name())
			}
			backendStack, ok := backends.GetProvider(dbConfig.Backend)
			if !ok {
				return gl.Errorf("Backend '%s' not found", dbConfig.Backend)
			}
			capabilities, err := backendStack.Capabilities(ctx)
			if err != nil {
				return gl.Errorf("Failed to get capabilities: %v", err)
			}
			if !capabilities.Managed {
				gl.Info("Backend is not managed. Skipping migration steps.")
				return nil
			}
			if hasMigrations, ok := capabilities.Features["migrations"]; !ok || !hasMigrations {
				gl.Info("Backend does not support migrations. Skipping migration steps.")
				return nil
			}

			// ========== STEP 3: GET MIGRATABLE BACKEND STACK ==========
			migratableStack, ok := backendStack.(provider.MigratableProvider)
			if !ok {
				return gl.Errorf("Backend '%s' does not support migrations", dbConfig.Backend)
			}
			endpoints, err := migratableStack.Start(ctx, provider.ConvertRootConfigToStartSpec(&rootConfig))
			if err != nil {
				return gl.Errorf("Failed to start services: %v", err)
			}
			gl.Info("Services started successfully.")

			// ========== STEP 4: RUN MIGRATIONS ==========
			mgr := engine.NewDatabaseManager(logger)
			for _, endpoint := range endpoints {
				if endpoint.DBConfig.Migration == nil {
					continue
				}
				conn, err := mgr.LoadDBConfig(endpoint.DBConfig)
				if err != nil {
					return gl.Errorf("Failed to get connection: %v", err)
				}
				if err := migratableStack.PrepareMigrations(ctx, &conn); err != nil {
					return gl.Errorf("Failed to prepare migrations: %v", err)
				}
				if err := migratableStack.RunMigrations(ctx, &conn, endpoint.DBConfig.Migration); err != nil {
					return gl.Errorf("Failed to migrate: %v", err)
				}
			}

			// ========== STEP 7 (OPTIONAL): ENGINE CONNECTIONS ==========
			if keepAlive {
				gl.Info("Establishing engine connections (keep-alive mode)...")
				mgr := engine.NewDatabaseManager(logger)
				if err := mgr.InitFromRootConfig(ctx, &rootConfig); err != nil {
					return gl.Errorf("Failed to initialize engine: %v", err)
				}
				gl.Info("Engine ready for runtime operations")
				// Note: In keep-alive mode, connections remain open.
				// Add graceful shutdown handling if needed.
			}

			gl.Info("Migration pipeline completed successfully!")

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVarP(&initArgs.DBConfigID, "target-db", "t", "", "Target database config ID")
	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().BoolVarP(&keepAlive, "keep-alive", "k", false, "Keep engine connections alive after migration (default: false)")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "", "Path to configuration file")

	// Future flags (not yet implemented)
	cmd.Flags().BoolVarP(&initArgs.Force, "force", "f", false, "Force apply all migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.Reset, "reset", "r", false, "Reset database before migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.DryRun, "dry-run", "", false, "Simulate migrations without applying (not yet implemented)")

	return cmd
}
