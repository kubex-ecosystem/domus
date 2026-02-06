package cli

import (
	"github.com/kubex-ecosystem/domus/internal/engine"

	dockerStack "github.com/kubex-ecosystem/domus/internal/backends/dockerstack"
	kbxInfo "github.com/kubex-ecosystem/kbx/tools/info"
	gl "github.com/kubex-ecosystem/logz"

	"context"
	"os"

	"github.com/kubex-ecosystem/domus/internal/module/info"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/services/docker"
	systemservice "github.com/kubex-ecosystem/domus/internal/services/system_service"
	"github.com/spf13/cobra"
)

var logger = gl.GetLoggerZ("Migration")

func DatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Database management commands for CanalizeDB"
	longDesc := "Database management commands for CanalizeDB"
	cmd := &cobra.Command{
		Use:         "database",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if initArgs.Debug {
				gl.SetDebugMode(initArgs.Debug)
			}
			gl.Info("CanalizeDB", "Database management commands for CanalizeDB")
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	cmd.AddCommand(startDatabaseCmd())
	cmd.AddCommand(stopDatabaseCmd())
	cmd.AddCommand(statusDatabaseCmd())
	cmd.AddCommand(migrateDatabaseCmd())

	return cmd
}

func startDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Start Database services"
	longDesc := "Start Database services. This will launch the database in a Docker container, and keep a minimal Z"

	cmd := &cobra.Command{
		Use:         "start",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(initArgs.Debug)
			if err := migrateDatabaseCmd().Execute(); err != nil {
				return gl.Errorf("Error executing migration: %v", err)
			}

			if err := systemservice.StartSystemService(initArgs); err != nil {
				return gl.Errorf("Error starting database services: %v", err)
			}
			gl.Info("CanalizeDB", "Database services started successfully.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")
	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

func stopDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Stop Docker"
	longDesc := "Stop Docker service"

	cmd := &cobra.Command{
		Use:         "stop",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

func statusDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}

	shortDesc := "Status Docker"
	longDesc := "Status Docker service"

	cmd := &cobra.Command{
		Use:         "status",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				gl.Errorf("Error displaying help: %v", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

	return cmd
}

// 	cmd := &cobra.Command{
// 		Use:         "service",
// 		Short:       shortDesc,
// 		Long:        longDesc,
// 		Aliases:     []string{"svc", "dbs"},
// 		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			gl.SetDebugMode(initArgs.Debug)

// 			if err := migrateDatabaseCmd().Execute(); err != nil {
// 				return gl.Errorf("Error executing migration: %v", err)
// 			}

// 			execPath, err := os.Executable()
// 			if err != nil {
// 				return gl.Errorf("Error getting executable path: %v", err)
// 			}
// 			gl.Infof("Executable path: %s", execPath)

// 			command := exec.Command(execPath, "database", "start", "--config-file", initArgs.ConfigFile)
// 			// command.Stdout = os.Stdout
// 			// command.Stderr = os.Stderr

// 			// Se pegarmos a saída não conseguimos spawnar em background, não msm
// 			proc := command.Process
// 			if proc == nil {
// 				return gl.Errorf("Failed to start database service: process is nil")
// 			}
// 			if proc.Pid == 0 {
// 				return gl.Errorf("Failed to start database service: invalid PID")
// 			}
// 			if err := proc.Release(); err != nil {
// 				return gl.Errorf("Failed to release process: %v", err)
// 			}

// 			gl.Infof("Starting database service with command: %s", command)

// 			// Note: The actual implementation to start the service in the background
// 			// may vary based on the operating system and requirements.
// 			// Here, we just log the command for demonstration purposes.

// 			return nil
// 		},
// 	}
// 	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")

// 	cmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", "", "Path to .env file")
// 	cmd.Flags().StringVar(&initArgs.ConfigFile, "config-file", "config.yaml", "Path to configuration file")

// 	return cmd
// }

func migrateDatabaseCmd() *cobra.Command {
	var initArgs = &kbx.InitArgs{}
	var keepAlive bool

	shortDesc := "Run database migrations"
	longDesc := "Run database migrations for all registered models."

	cmd := &cobra.Command{
		Use:         "migrate",
		Short:       shortDesc,
		Long:        longDesc,
		Annotations: kbxInfo.CLIBannerStyle(info.GetBanners(), []string{shortDesc, longDesc}, (os.Getenv("DOMUS_HIDEBANNER") == "true")),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize context and logger
			ctx := context.Background()
			gl.SetDebugMode(initArgs.Debug)

			// ========== STEP 1: LOAD CONFIG ==========
			gl.Info("Loading configuration...")
			rootConfig, err := engine.LoadRootConfig(
				kbx.GetValueOrDefaultSimple(initArgs.ConfigFile, os.ExpandEnv(kbx.DefaultConfigFile)),
			)
			if err != nil {
				return gl.Errorf("Failed to load config: %v", err)
			}

			// ========== STEP 2: CREATE SERVICE MANAGER ==========
			gl.Info("Initializing Docker service...")
			dockerService, err := docker.NewDockerService(logger)
			if err != nil {
				return gl.Errorf("Failed to create Docker service: %v", err)
			}

			// ========== STEP 3: CREATE PROVIDER WITH INJECTION ==========

			gl.Info("Initializing DockerStack provider...")
			dsp := dockerStack.NewDockerStackProvider(dockerService)

			// ========== STEP 4-6: PROVIDER ORCHESTRATES EVERYTHING ==========
			gl.Info("Starting migration pipeline...")
			if err := dsp.StartServices(ctx, &rootConfig); err != nil {
				return gl.Errorf("Migration pipeline failed: %v", err)
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
	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Enable debug mode")
	cmd.Flags().BoolVarP(&keepAlive, "keep-alive", "k", false, "Keep engine connections alive after migration (default: false)")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "C", "config.yaml", "Path to configuration file")

	// Future flags (not yet implemented)
	cmd.Flags().BoolVarP(&initArgs.Force, "force", "f", false, "Force apply all migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.Reset, "reset", "r", false, "Reset database before migrations (not yet implemented)")
	cmd.Flags().BoolVarP(&initArgs.DryRun, "dry-run", "", false, "Simulate migrations without applying (not yet implemented)")

	return cmd
}
