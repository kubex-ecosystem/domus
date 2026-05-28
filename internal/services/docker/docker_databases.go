// Package docker implementa serviços Docker para bancos de dados.
package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
	ci "github.com/kubex-ecosystem/domus/internal/interfaces"
	"github.com/kubex-ecosystem/domus/internal/module/kbx"
	svc "github.com/kubex-ecosystem/domus/internal/services"
	kbxCrp "github.com/kubex-ecosystem/ethyr/tools/security/crypto"
	kbxKrs "github.com/kubex-ecosystem/ethyr/tools/security/external"
	logz "github.com/kubex-ecosystem/logz"
)

// getMongoDatabaseSetup inicia o serviço MongoDB
func getMongoDatabaseSetup(ctx context.Context, d ci.IDockerService, services []*ci.Services, dbConfig *kbx.DBConfig) ([]*ci.Services, error) {
	// Check if the database is already running
	if ok := IsServiceRunning("kubexdb-mongo"); ok {
		logz.Log("debug", fmt.Sprintf("%s já está rodando!", "kubexdb-mongo"))
		return services, nil
	}

	logz.Notice("Iniciando serviço MongoDB")
	if err := d.StartContainerByName("kubexdb-mongo"); err == nil {
		logz.Log("debug", fmt.Sprintf("%s já está rodando!", "kubexdb-mongo"))
		return services, nil
	}

	// Check if Password is empty, if so, try to retrieve it from keyring
	// if not found, generate a new one
	if dbConfig.Pass == "" {
		logz.Debug("Password not found in config, generating new one and storing in filekeyring")
		kbxKrsSvc := kbxKrs.NewFileKeyringService("domus", "mongodb")
		kbxCrpSvc := kbxCrp.NewCryptoService()
		mgoPassKey, mgoPassErr := kbxCrpSvc.GenerateKeyWithLength(16)
		if mgoPassErr != nil {
			logz.Log("error", fmt.Sprintf("Error generating key: %v", mgoPassErr))
			return services, mgoPassErr
		}
		mgoPassErr = kbxKrsSvc.StorePassword(string(mgoPassKey))
		if mgoPassErr != nil {
			logz.Log("error", fmt.Sprintf("Error saving key to keyring: %v", mgoPassErr))
			return services, mgoPassErr
		}
		dbConfig.Pass = string(mgoPassKey)
	} else {
		logz.Log("debug", fmt.Sprintf("Password found in config: %s", dbConfig.Pass[0:2]))
	}

	var vol string
	if volume, ok := dbConfig.Options["volume"]; ok {
		vol = volume.(string)
	}

	if len(vol) == 0 {
		vol = os.ExpandEnv(kbx.DefaultMongoVolume)
	} else {
		vol = os.ExpandEnv(vol)
	}

	mgoVolRootDir := os.ExpandEnv(vol)
	mgoVolInitDir := os.ExpandEnv(filepath.Join(kbx.DefaultMongoVolume, "init"))
	volsTmp := make(map[string]string)
	if strings.Contains(mgoVolRootDir, ":") {
		parts := strings.SplitN(mgoVolRootDir, ":", 2)
		hostPath := parts[0]
		containerPath := parts[1]
		volsTmp[hostPath] = containerPath
	} else {
		volsTmp[mgoVolRootDir] = "/data/db"
	}
	vols := make(map[string]struct{})
	for hostPath, containerPath := range volsTmp {
		vols[strings.Join([]string{hostPath, containerPath}, ":")] = struct{}{}
	}
	// insert init dir volume
	vols[strings.Join([]string{mgoVolInitDir, "/docker-entrypoint-initdb.d"}, ":")] = struct{}{}
	counter := 0
	for volName := range volsTmp {
		if err := d.CreateVolume("kubexdb-mongo-vol-"+fmt.Sprint(counter), volName); err != nil {
			logz.Log("error", fmt.Sprintf("Erro ao criar volume do MongoDB: %v", err))
			return services, err
		}
		counter++
	}

	// Check if the port is already in use and find an available one if necessary
	if (dbConfig.Port == "") || len(dbConfig.Port) == 0 {
		dbConfig.Port = "27017"
	}

	nPort, _ := strconv.Atoi(dbConfig.Port)
	port, err := svc.FindAvailablePort(nPort, 10)
	if err != nil {
		logz.Log("error", fmt.Sprintf("Erro ao encontrar porta disponível: %v", err))
		return services, err
	}
	dbConfig.Port = port
	// Map the port to the container
	portMap := d.MapPorts(dbConfig.Port, "27017/tcp")

	// Check if the database name is empty, if so, generate a random one
	if dbConfig.Name == "" {
		dbConfig.Name = "domus-" + svc.RandStringBytes(5)
	}
	logz.Log("debug", fmt.Sprintf("Iniciando MongoDB com o nome: %s", dbConfig.Name))

	// Insert the MongoDB service into the services slice
	dbConnObj := NewServices(
		"kubexdb-mongo",
		"mongo:7.0",
		map[string]string{
			// Standard configs for Mongo container
			"MONGO_INITDB_ROOT_USERNAME": dbConfig.User,
			"MONGO_INITDB_ROOT_PASSWORD": dbConfig.Pass,
			"MONGO_INITDB_DATABASE":      dbConfig.DBName,
			"MONGO_DATABASE":             dbConfig.DBName,
			"MONGO_USERNAME":             dbConfig.User,
			"MONGO_PASSWORD":             dbConfig.Pass,
			"MONGO_PORT":                 dbConfig.Port,
			"MONGO_HOST":                 dbConfig.Host,
			"MONGO_DATA":                 "/data/db",
			"MONGO_SSL":                  "no",
			"MONGO_ENABLE_TLS":           "no",
			"MONGO_TLS_MODE":             "disable",
			"MONGO_TLS_CERT_FILE":        "",
			"MONGO_TLS_KEY_FILE":         "",
			"MONGO_TLS_CA_FILE":          "",
			"MONGO_DATA_CHECKSUMS":       "on",
			"MONGO_LOGGING_COLLECTOR":    "on",
			"MONGO_LOGGING_COLORS":       "on",
			"MONGO_LOG_STATEMENT":        "all",
			"MONGO_POOL_SIZE":            "50",
			"MONGO_MAX_CONNECTIONS":      "200",
			"MONGO_INITDB_WALDIR":        "/data/db/wal",
			"MONGO_HOST_AUTH_METHOD":     "md5",
			"MONGO_ALLOW_EMPTY_PASSWORD": "no",
		},
		[]nat.PortMap{portMap},
		vols,
		[]string{
			"mongod",
			"--config",
			"/etc/mongodb/mongodb.conf",
		},
	)
	services = append(services, dbConnObj)

	return services, nil
}

// getPGDatabaseSetup gets the PG database setup
func getPGDatabaseSetup(ctx context.Context, d ci.IDockerService, services []*ci.Services, dbConfig *kbx.DBConfig) ([]*ci.Services, error) {
	// Check if the database is already running
	ok := IsServiceRunning("kubexdb-pg")
	if ok {
		logz.Log("debug", fmt.Sprintf("%s já está rodando!", "kubexdb-pg"))
		return services, nil
	}

	if err := d.StartContainerByName("kubexdb-pg"); err == nil {
		logz.Log("debug", fmt.Sprintf("%s já está rodando!", "kubexdb-pg"))
		return services, nil
	}

	// Check if Password is empty, if so, try to retrieve it from keyring
	// if not found, generate a new one
	if dbConfig.Pass == "" {
		logz.Debug("Password not found in config, generating new one and storing in filekeyring")
		kbxKrsSvc := kbxKrs.NewFileKeyringService("domus", "postgres")
		kbxCrpSvc := kbxCrp.NewCryptoService()
		pgPassKey, pgPassErr := kbxCrpSvc.GenerateKeyWithLength(16)
		if pgPassErr != nil {
			logz.Log("error", fmt.Sprintf("Error generating key: %v", pgPassErr))
			return services, pgPassErr
		}
		pgPassErr = kbxKrsSvc.StorePassword(string(pgPassKey))
		if pgPassErr != nil {
			logz.Log("error", fmt.Sprintf("Error saving key to keyring: %v", pgPassErr))
			return services, pgPassErr
		}
		dbConfig.Pass = string(pgPassKey)
	} else {
		logz.Log("debug", fmt.Sprintf("Password found in config: %s", dbConfig.Pass[0:2]))
	}

	var vol string
	if volume, ok := dbConfig.Options["volume"]; ok {
		vol = volume.(string)
	}

	if len(vol) == 0 {
		vol = os.ExpandEnv(kbx.DefaultPostgresVolume)
	} else {
		vol = os.ExpandEnv(vol)
	}

	pgVolRootDir := os.ExpandEnv(vol)
	pgVolInitDir := os.ExpandEnv(filepath.Join(kbx.DefaultPostgresVolume, "init"))
	volsTmp := make(map[string]string)
	if strings.Contains(pgVolRootDir, ":") {
		parts := strings.SplitN(pgVolRootDir, ":", 2)
		hostPath := parts[0]
		containerPath := parts[1]
		volsTmp[hostPath] = containerPath
	} else {
		volsTmp[pgVolRootDir] = "/var/lib/postgresql/data"
	}
	vols := make(map[string]struct{})
	for hostPath, containerPath := range volsTmp {
		vols[strings.Join([]string{hostPath, containerPath}, ":")] = struct{}{}
	}
	// insert init dir volume
	vols[strings.Join([]string{pgVolInitDir, "/docker-entrypoint-initdb.d"}, ":")] = struct{}{}
	counter := 0
	for volName := range volsTmp {
		if err := d.CreateVolume("kubexdb-pg-vol-"+fmt.Sprint(counter), volName); err != nil {
			logz.Log("error", fmt.Sprintf("Erro ao criar volume do PostgreSQL: %v", err))
			return services, err
		}
		counter++
	}

	// Check if the port is already in use and find an available one if necessary
	if (dbConfig.Port == "") || len(dbConfig.Port) == 0 {
		dbConfig.Port = "5432"
	}

	nPort, _ := strconv.Atoi(dbConfig.Port)
	port, err := svc.FindAvailablePort(nPort, 10)
	if err != nil {
		logz.Log("error", fmt.Sprintf("Erro ao encontrar porta disponível: %v", err))
		return services, err
	}
	dbConfig.Port = port
	// Map the port to the container
	portMap := d.MapPorts(dbConfig.Port, "5432/tcp")

	// Check if the database name is empty, if so, generate a random one
	if dbConfig.Name == "" {
		dbConfig.Name = "domus-" + svc.RandStringBytes(5)
	}
	enableTLS := func(dbConfig *kbx.DBConfig) string {
		if dbConfig.TLSEnabled {
			return "require"
		}
		return "disable"
	}

	// Insert the PostgreSQL service into the services slice
	dbConnObj := NewServices(
		"kubexdb-pg",
		"postgres:17-alpine",
		map[string]string{
			// Standard configs for Postgres container
			"POSTGRES_APPLICATION_NAME": dbConfig.Name,
			// "POSTGRES_SERVICE_NAME":      "domus_service",
			"POSTGRES_INITDB_ARGS":       "--encoding=UTF8 --locale=pt_BR.UTF-8 --data-checksums",
			"POSTGRES_USER":              dbConfig.User,
			"POSTGRES_PASSWORD":          dbConfig.Pass,
			"POSTGRES_DB":                dbConfig.DBName,
			"POSTGRES_PORT":              dbConfig.Port,
			"POSTGRES_DB_NAME":           dbConfig.DBName,
			"POSTGRES_DB_ENCODING":       "UTF8",
			"POSTGRES_LOGGING_COLLECTOR": "on",
			"POSTGRES_LOGGING_COLORS":    "on",
			"POSTGRES_LOG_STATEMENT":     "all",
			"POSTGRES_POOL_SIZE":         "50",
			"POSTGRES_MAX_CONNECTIONS":   "200",

			// Necessary for some clients
			"PGAPPNAME": dbConfig.Name,
			// "PGSERVICE":              "domus_service",
			"PGUSER":     dbConfig.User,
			"PGPASSWORD": dbConfig.Pass,
			"PGDATABASE": dbConfig.DBName,
			"PGPORT":     dbConfig.Port,
			"PGHOST":     dbConfig.Host,
			"PGDATA":     "/var/lib/postgresql/data/pgdata",
			"PGSSLMODE":  enableTLS(dbConfig),
			// "TZ":                     "UTC",
			// "LANG":                   "pt_BR.utf8",
			// "LC_ALL":                 "pt_BR.utf8",
			// "LANGUAGE":               "pt_BR:en",
			"PGCLIENTENCODING": "UTF8",
			// "PGDATESTYLE":            "ISO, MDY",
			// "PGTIMEZONE":             "UTC",
			// "PGDATACHECKSUMS":        "on",
			"POSTGRES_INITDB_WALDIR": "/var/lib/postgresql/data/pg_wal",

			// Development configs
			// Outra opção sem ser trust, com essa senha dbConfig.Pass
			// "POSTGRES_HOST_AUTH_METHOD": "md5",
			"PG_ALLOW_EMPTY_PASSWORD": "no",
			"POSTGRES_ENABLE_TLS":     "no",
			// "POSTGRES_TLS_MODE":         "disable",
			// "POSTGRES_TLS_CERT_FILE":    "",
			// "POSTGRES_TLS_KEY_FILE":     "",
			// "POSTGRES_TLS_CA_FILE":      "",

			// To avoid interactive dialog during installation
			"DEBIAN_FRONTEND": "noninteractive",
		},
		[]nat.PortMap{portMap},
		vols,
		[]string{
			"postgres",
			"-c", "shared_buffers=1GB",
			"-c", "work_mem=16MB",
			"-c", "maintenance_work_mem=256MB",
			"-c", "effective_cache_size=2GB",
			"-c", "max_connections=120",
			"-c", "synchronous_commit=off",
			"-c", "checkpoint_timeout=10min",
			"-c", "checkpoint_completion_target=0.9",
			"-c", "max_wal_size=2GB",
			"-c", "min_wal_size=256MB",
			"-c", "random_page_cost=2.0",
			"-c", "effective_io_concurrency=50",
			"-c", "autovacuum=on",
			"-c", "autovacuum_max_workers=3",
			"-c", "autovacuum_naptime=1min",
			"-c", "log_min_duration_statement=300ms",
			"-c", "log_checkpoints=on",
			"-c", "log_lock_waits=on",
			// Set password encryption method
			"-c", "password_encryption=scram-sha-256",
			// Enable SSL if required
			"-c", "ssl=off",
			// "-c", "ssl=on",
		},
	)
	services = append(services, dbConnObj)

	return services, nil
}

// getRedisDatabaseSetup gets the Redis database setup
func getRedisDatabaseSetup(ctx context.Context, d ci.IDockerService, services []*ci.Services, dbConfig *kbx.DBConfig) ([]*ci.Services, error) {
	// Check if the database is already running
	if IsServiceRunning(dbConfig.Name) {
		logz.Log("info", fmt.Sprintf("Serviço %s já está rodando!", dbConfig.Name))
		return services, nil
	}

	// Create a map for port mappings
	portMap := d.MapPorts(dbConfig.Port, "6379/tcp")

	// Check if the database name is empty, if so, generate a random one
	if dbConfig.Name == "" {
		dbConfig.Name = "kubexdb-redis-" + svc.RandStringBytes(5)
	}

	// Insert the PostgreSQL service into the services slice
	dbConnObj := NewServices(
		dbConfig.Name,
		"redis:7-alpine",
		map[string]string{
			"REDIS_PORT":     dbConfig.Port,
			"REDIS_PASSWORD": dbConfig.Pass,
			"REDIS_DB":       dbConfig.DBName,
		},
		[]nat.PortMap{portMap},
		nil,
		[]string{
			"redis-server",
			"--appendonly", "yes",
			"--save", "",
			"--requirepass", dbConfig.Pass,
		},
	)
	services = append(services, dbConnObj)

	return services, nil
}

// getHTTPDatabaseSetup gets the HTTP database setup
func getHTTPDatabaseSetup(ctx context.Context, d ci.IDockerService, services []*ci.Services, dbConfig *kbx.DBConfig) ([]*ci.Services, error) {
	// Check if the database is already running
	if IsServiceRunning(dbConfig.Name) {
		logz.Log("info", fmt.Sprintf("Serviço %s já está rodando!", dbConfig.Name))
		return services, nil
	}

	// Create a map for port mappings
	portMap := d.MapPorts(dbConfig.Port, "80/tcp")

	// Check if the database name is empty, if so, generate a random one
	if dbConfig.Name == "" {
		dbConfig.Name = "kubexdb-http-" + svc.RandStringBytes(5)
	}

	// Insert the HTTP service into the services slice
	dbConnObj := NewServices(
		dbConfig.Name,
		"nginx:latest",
		map[string]string{
			"NGINX_PORT": dbConfig.Port,
		},
		[]nat.PortMap{portMap},
		nil,
		[]string{
			"nginx",
		},
	)
	services = append(services, dbConnObj)

	return services, nil
}

// SetupDatabaseServices inicia os serviços de banco de dados
func SetupDatabaseServices(ctx context.Context, d ci.IDockerService, rootConfig *kbx.RootConfig) error {
	if rootConfig == nil {
		return fmt.Errorf("Configuração do banco de dados não encontrada")
	}
	if !kbx.DefaultTrue(rootConfig.Enabled) {
		logz.Log("debug", "Database services are disabled in config, skipping DB setup")
		return nil
	}
	if len(rootConfig.Databases) == 0 {
		logz.Log("debug", "Not found connections in config, skipping DB setup")
		return nil
	}
	var services = make([]*ci.Services, 0)
	var err error

	// START GENERIC DATABASE CONFIGS
	if len(rootConfig.Databases) > 0 {
		for _, dbConfig := range rootConfig.Databases {
			if (dbConfig.Protocol == "postgres" || dbConfig.Protocol == "postgresql") && kbx.DefaultTrue(dbConfig.Enabled) {
				services, err = getPGDatabaseSetup(ctx, d, services, &dbConfig)
				if err != nil {
					return err
				}
			} else if dbConfig.Protocol == "mongo" && kbx.DefaultTrue(dbConfig.Enabled) {
				services, err = getMongoDatabaseSetup(ctx, d, services, &dbConfig)
				if err != nil {
					return err
				}
			}
			if (dbConfig.Protocol == "redis") && kbx.DefaultTrue(dbConfig.Enabled) {
				services, err = getRedisDatabaseSetup(ctx, d, services, &dbConfig)
				if err != nil {
					return err
				}
			}
			if (dbConfig.Protocol == "http") && kbx.DefaultTrue(dbConfig.Enabled) {
				services, err = getHTTPDatabaseSetup(ctx, d, services, &dbConfig)
				if err != nil {
					return err
				}
			}
		}
	} else {
		logz.Log("debug", "Not found databases in config, skipping DB setup")
	}

	logz.Log("debug", fmt.Sprintf("Iniciando %d serviços...", len(services)))
	for _, srv := range services {
		mapPorts := map[nat.Port]struct{}{}
		for _, port := range srv.Ports {
			pt := svc.ExtractPort(port)
			if pt == "" {
				logz.Log("error", fmt.Sprintf("Erro ao mapear porta %s", pt))
				continue
			}
			if _, ok := pt.(map[string]string); !ok {
				logz.Log("error", fmt.Sprintf("Erro ao mapear porta %s", pt))
				continue
			}
			// Verifica se a porta já está mapeada
			ptStr, ok := pt.(map[string]string)
			if !ok || ptStr["port"] == "" || ptStr["protocol"] == "" {
				logz.Log("error", fmt.Sprintf("Erro ao mapear porta: tipo inválido ou campos ausentes: %v", pt))
				continue
			}
			portKey := nat.Port(fmt.Sprintf("%s/%s", ptStr["port"], ptStr["protocol"]))
			if _, exists := mapPorts[portKey]; exists {
				logz.Log("error", fmt.Sprintf("Erro ao mapear porta %s", portKey))
				continue
			}
			// Adiciona a porta ao mapa
			portMap, ok := pt.(map[string]string)
			if !ok {
				logz.Log("error", fmt.Sprintf("Erro ao converter porta: %v", pt))
				continue
			}
			portKey = nat.Port(fmt.Sprintf("%s/%s", portMap["port"], portMap["protocol"]))
			mapPorts[portKey] = struct{}{}
		}
		// Verifica se o serviço já está rodando
		// Isso já está dentro do StartContainer
		// if IsServiceRunning(srv.Name) {
		// 	logz.Log("info", fmt.Sprintf("%s já está rodando!", srv.Name))
		// 	continue
		// }
		if err := d.StartContainer(srv.Name, srv.Image, srv.Env, mapPorts, srv.Volumes, srv.Cmd); err != nil {
			return err
		}
	}
	return nil
}
