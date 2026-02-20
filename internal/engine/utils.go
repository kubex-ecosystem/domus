package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	"github.com/kubex-ecosystem/domus/internal/types"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	logz "github.com/kubex-ecosystem/logz"
)

// LoadRootConfig carrega um arquivo JSON simples de config.
func LoadRootConfig(path string) (kbx.RootConfig, error) {
	if path == "" {
		path = os.ExpandEnv(kbxGet.EnvOr("KUBEX_DOMUS_CONFIG_PATH", kbx.DefaultKubexDomusConfigPath))
	}

	if _, err := os.Stat(path); err != nil {
		logz.Fatalf("Failed to load DS RootConfig %s: %v", path, err)
		// return kbx.RootConfig{}, err
	} else if errors.Is(err, os.ErrNotExist) {
		logz.Infof("Config file %s does not exist, generating default config", path)
		defaultCfg := GenerateDefaultPostgresConfig()
		defaultCfg.FilePath = path
		if err := SaveRootConfig(&defaultCfg); err != nil {
			return kbx.RootConfig{}, fmt.Errorf("failed to save default config to %s: %v", path, err)
		}
		logz.Infof("Default config saved to %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return kbx.RootConfig{}, err
	}
	// var cfg *RootConfig
	cfgMp := types.NewMapperType(&kbx.RootConfig{}, path)
	cfgObj, err := cfgMp.Deserialize(data, filepath.Ext(path)[1:])
	if err != nil {
		return kbx.RootConfig{}, err
	}
	if cfgObj != nil {
		return *cfgObj, nil
	}

	newPath := filepath.Join(os.ExpandEnv(kbx.DefaultConfigDir), "domus", "config", filepath.Base(path))
	cfgMpC := types.NewMapperType(cfgMp.GetObject(), os.ExpandEnv(newPath))
	cfgMpC.SerializeToFile(filepath.Ext(path)[1:])
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return kbx.RootConfig{}, fmt.Errorf("config file not found at %s", path)
	}

	return kbx.RootConfig{}, errors.New("failed to deserialize root config")
}

// SaveRootConfig salva o arquivo JSON.
func SaveRootConfig(cfg *kbx.RootConfig) error {
	if cfg.FilePath == "" {
		return errors.New("root config FilePath is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.FilePath, data, 0o640)
}

// GetDefaultConfigPath calcula o path padrão $HOME/.gnyx/database/postgres/config.json
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(
		home,
		".gnyx",
		"database",
		"domus",
		"config.json",
	), nil
}

// GenerateRandomPassword é só um helper simples (pode trocar pela sua versão oficial).
func GenerateRandomPassword(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	buf := make([]byte, n)
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// fallback tosco, mas ok
		for i := range buf {
			buf[i] = alphabet[i%len(alphabet)]
		}
		return string(buf)
	}
	defer f.Close()
	_, _ = f.Read(buf)
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf)
}

// GenerateDefaultPostgresConfig gera uma única config de Postgres básica.
func GenerateDefaultPostgresConfig() kbx.RootConfig {
	pass := GenerateRandomPassword(40)

	db := kbx.DBConfig{
		// ID:        "postgres",
		Name:      kbxGet.EnvOr("KUBEX_DOMUS_DB_NAME", "domus"),
		IsDefault: true,
		Enabled: kbxGet.ValueOrIf(
			kbxGet.EnvOr("KUBEX_DOMUS_DB_ENABLED", "") != "",
			kbxGet.EnvOrType(
				"KUBEX_DOMUS_DB_ENABLED",
				kbxGet.BlPtr(true),
			),
			kbxGet.BlPtr(true),
		),
		// Type:      DBTypePostgres,
		Host:   kbxGet.EnvOr("KUBEX_DOMUS_DB_HOST", "127.0.0.1"),
		Port:   kbxGet.EnvOr("KUBEX_DOMUS_DB_PORT", "5432"),
		User:   kbxGet.EnvOr("KUBEX_DOMUS_DB_USER", "kubex_adm"),
		Pass:   pass,
		DBName: kbxGet.EnvOr("KUBEX_DOMUS_DB_NAME", "domus"),
		Schema: kbxGet.EnvOr("KUBEX_DOMUS_DB_SCHEMA", "public"),
		Options: map[string]any{
			"sslmode":           kbxGet.EnvOr("KUBEX_DOMUS_DB_SSLMODE", "disable"),
			"max_connections":   kbxGet.EnvOrType("KUBEX_DOMUS_DB_MAX_CONNECTIONS", 50),
			"connect_timeout":   kbxGet.EnvOrType("KUBEX_DOMUS_DB_CONNECT_TIMEOUT", 10),
			"application_name":  kbxGet.EnvOr("KUBEX_DOMUS_DB_APPLICATION_NAME", "domus"),
			"pool_max_lifetime": kbxGet.EnvOr("KUBEX_DOMUS_DB_POOL_MAX_LIFETIME", "30m"),
		},
	}

	// // DSN simples, o driver pode refinar
	// db.DSN =

	dsn := types.NewDSN(
		kbxGet.StrPtr(
			fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=%s",
				db.User,
				db.Pass,
				db.Host,
				db.Port,
				db.DBName,
				db.Options["sslmode"],
			),
		),
	)

	if err := dsn.Validate(); err != nil {
		return kbx.RootConfig{} //, fmt.Errorf("failed to validate DSN: %s", dsn.Redact())
	}

	return kbx.RootConfig{
		Name:      "domus",
		Enabled:   kbx.BoolPtr(true),
		Databases: []kbx.DBConfig{db},
	}
}

// BootstrapDatabaseManager é o entrypoint que o main do DS pode chamar.
func BootstrapDatabaseManager(ctx context.Context, logger *logz.LoggerZ, cfgPath string) (kbx.RootConfig, error) {
	mgr := NewDatabaseManager(logger)

	root, err := mgr.LoadOrBootstrap(cfgPath)
	if err != nil {
		return kbx.RootConfig{}, err
	}

	if err := mgr.InitFromRootConfig(ctx, &root); err != nil {
		return kbx.RootConfig{}, err
	}

	return root, nil
}
