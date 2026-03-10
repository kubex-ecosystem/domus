package integrationstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kubex-ecosystem/domus/internal/execution"
	t "github.com/kubex-ecosystem/domus/internal/types"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/kbx/tools/security/crypto"
)

// pgIntegrationStore implementa IntegrationStore usando PGExecutor e CryptoService.
type pgIntegrationStore struct {
	exec   execution.PGExecutor
	crypto *crypto.CryptoService
	dbKey  []byte // Master key injetada na inicialização do store
}

// NewPGIntegrationStore cria uma instância do store com criptografia embutida.
func NewPGIntegrationStore(exec execution.PGExecutor, masterKey []byte) IntegrationStore {
	return &pgIntegrationStore{
		exec:   exec,
		crypto: crypto.NewCryptoService(),
		dbKey:  masterKey,
	}
}

// ============================================================================
// MÉTODOS DE PERSISTÊNCIA (JOBS)
// ============================================================================

func (s *pgIntegrationStore) CreateJob(ctx context.Context, input *SyncJob) (*SyncJob, error) {
	if input == nil {
		return nil, fmt.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TaskName) == "" {
		return nil, fmt.Errorf("task name is required: %v", t.ErrInvalidInput)
	}

	const q = `
		INSERT INTO sync_job (
			tenant_id, config_id, task_name, cron_expression, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tenant_id, config_id, task_name, cron_expression, is_active, last_sync_at, created_at, updated_at
	`

	now := time.Now().UTC()
	row := s.exec.QueryRow(ctx, q,
		input.TenantID,
		input.ConfigID,
		input.TaskName,
		input.CronExpression,
		input.IsActive,
		now,
	)

	job, err := scanSyncJob(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync job: %v", err)
	}

	return job, nil
}

func (s *pgIntegrationStore) ListActiveJobs(ctx context.Context, tenantID string) ([]SyncJob, error) {
	const q = `
		SELECT id, tenant_id, config_id, task_name, cron_expression, is_active, last_sync_at, created_at, updated_at
		FROM sync_job
		WHERE tenant_id = $1 AND is_active = true
	`

	rows, err := s.exec.Query(ctx, q, tenantID)
	if err != nil {
		return nil, gl.Errorf("failed to list active jobs: %v", err)
	}
	defer rows.Close()

	var jobs []SyncJob
	for rows.Next() {
		job, err := scanSyncJob(rows)
		if err != nil {
			return nil, gl.Errorf("failed to scan sync job: %v", err)
		}
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

func (s *pgIntegrationStore) UpdateJobLastSync(ctx context.Context, jobID string, syncTime time.Time) error {
	const q = `
		UPDATE sync_job
		SET last_sync_at = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := s.exec.Exec(ctx, q, syncTime, time.Now().UTC(), jobID)
	if err != nil {
		return gl.Errorf("failed to update job last sync time: %v", err)
	}
	return nil
}

// ============================================================================
// MÉTODOS DE PERSISTÊNCIA (CONFIGS)
// ============================================================================

// CreateConfig insere uma nova configuração, blindando os dados sensíveis do JSONB.
func (s *pgIntegrationStore) CreateConfig(ctx context.Context, input *IntegrationConfig) (*IntegrationConfig, error) {
	if input == nil {
		return nil, gl.Errorf("create input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, gl.Errorf("name is required: %v", t.ErrInvalidInput)
	}

	// 1. Intercepta e criptografa as chaves sensíveis do JSON
	secureSettings, err := s.encryptSensitiveFields(input.Settings)
	if err != nil {
		return nil, gl.Errorf("failed to encrypt settings: %v", err)
	}

	const q = `
		INSERT INTO integration_config (
			tenant_id, partner_id, type, name, settings, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
	`

	now := time.Now().UTC()
	row := s.exec.QueryRow(ctx, q,
		input.TenantID,
		input.PartnerID,
		input.Type,
		input.Name,
		secureSettings,
		input.IsActive,
		now,
	)

	config, err := scanIntegrationConfig(row)
	if err != nil {
		return nil, gl.Errorf("failed to create integration config: %v", err)
	}

	// 2. Descriptografa antes de devolver para a camada em memória
	config.Settings, _ = s.decryptSensitiveFields(config.Settings)

	return config, nil
}

// GetConfigByID busca a configuração e abre o cofre do JSONB automaticamente.
func (s *pgIntegrationStore) GetConfigByID(ctx context.Context, id string) (*IntegrationConfig, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}

	const q = `
		SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
		FROM integration_config
		WHERE id = $1
	`

	row := s.exec.QueryRow(ctx, q, id)
	config, err := scanIntegrationConfig(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, gl.Errorf("failed to get integration config: %v", err)
	}

	// Abre o cofre antes de entregar pro Domus/GNyx
	config.Settings, err = s.decryptSensitiveFields(config.Settings)
	if err != nil {
		return nil, gl.Errorf("failed to decrypt settings: %v", err)
	}

	return config, nil
}

func (s *pgIntegrationStore) ListConfigsByTenant(ctx context.Context, tenantID string) ([]IntegrationConfig, error) {
	const q = `
		SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
		FROM integration_config
		WHERE tenant_id = $1
	`

	rows, err := s.exec.Query(ctx, q, tenantID)
	if err != nil {
		return nil, gl.Errorf("failed to list integration configs: %v", err)
	}
	defer rows.Close()

	var configs []IntegrationConfig
	for rows.Next() {
		config, err := scanIntegrationConfig(rows)
		if err != nil && config != nil {
			return nil, gl.Errorf("failed to scan integration config: %v", err)
		}
		configs = append(configs, *config)
	}

	return configs, nil
}

// ============================================================================
// HELPERS DE CRIPTOGRAFIA DINÂMICA
// ============================================================================

// encryptSensitiveFields varre o JSON procurando segredos e os criptografa.
func (s *pgIntegrationStore) encryptSensitiveFields(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	for key, value := range data {
		if isSensitiveKey(key) {
			if strVal, ok := value.(string); ok && strVal != "" {
				// Evita dupla encriptação
				if strings.HasPrefix(strVal, "enc_v1:") {
					continue
				}

				// Chama a interface do seu CryptoService
				// Assumindo que o primeiro retorno é a string encriptada (Base64/Hex) encodada com o nonce
				encryptedStr, _, err := s.crypto.Encrypt([]byte(strVal), s.dbKey)
				if err != nil {
					return nil, gl.Errorf("encryption failed for key %s: %w", key, err)
				}

				// Sela o valor
				data[key] = "enc_v1:" + encryptedStr
			}
		}
	}
	return json.Marshal(data)
}

// decryptSensitiveFields varre o JSON e restaura os valores originais.
func (s *pgIntegrationStore) decryptSensitiveFields(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	for key, value := range data {
		if strVal, ok := value.(string); ok && strings.HasPrefix(strVal, "enc_v1:") {
			// Remove o selo para pegar o payload real
			cipherText := strings.TrimPrefix(strVal, "enc_v1:")

			// Chama a interface do seu CryptoService
			// Recebe o payload e a master key, retorna a string decriptada no 1º output
			decryptedStr, _, err := s.crypto.Decrypt([]byte(cipherText), s.dbKey)
			if err != nil {
				return nil, gl.Errorf("decryption failed for key %s: %w", key, err)
			}

			// Devolve pro mapa
			data[key] = decryptedStr
		}
	}
	return json.Marshal(data)
}

// isSensitiveKey define o que deve ser selado dentro do JSONB.
func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "ssh") ||
		strings.Contains(k, "key")
}

// ============================================================================
// MÉTODOS BASICOS DE CRUD (EXEMPLO PARA CONFIGS, REPLICAR PARA JOBS SE NECESSÁRIO)
// ============================================================================

func (s *pgIntegrationStore) Create(ctx context.Context, config *IntegrationConfig) error {
	created, err := s.CreateConfig(ctx, config)
	if err != nil {
		return err
	}
	*config = *created
	return nil
}

func (s *pgIntegrationStore) GetByID(ctx context.Context, id string) (*IntegrationConfig, error) {
	const q = `
		SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
		FROM integration_config
		WHERE id = $1
	`

	row := s.exec.QueryRow(ctx, q, id)
	config, err := scanIntegrationConfig(row)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *pgIntegrationStore) GetBySlug(ctx context.Context, tenantID string, slug string) (*IntegrationConfig, error) {
	const q = `
		SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
		FROM integration_config
		WHERE tenant_id = $1 AND name = $2
	`

	row := s.exec.QueryRow(ctx, q, tenantID, slug)
	config, err := scanIntegrationConfig(row)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *pgIntegrationStore) Update(ctx context.Context, config *IntegrationConfig) error {
	const q = `
		UPDATE integration_config
		SET tenant_id = $2, partner_id = $3, type = $4, name = $5, settings = $6, is_active = $7, updated_at = NOW()
		WHERE id = $1
	`

	_, err := s.exec.Exec(ctx, q,
		config.ID,
		config.TenantID,
		config.PartnerID,
		config.Type,
		config.Name,
		config.Settings,
		config.IsActive)
	if err != nil {
		return err
	}
	return nil
}

func (s *pgIntegrationStore) Delete(ctx context.Context, id string) error {
	const q = `
		DELETE FROM integration_config
		WHERE id = $1
	`

	res, err := s.exec.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return gl.Errorf("no rows affected for ID %s", id)
	}
	return nil
}

func (s *pgIntegrationStore) List(ctx context.Context, tenantID string) ([]*IntegrationConfig, error) {
	const q = `
        SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
        FROM integration_config
        WHERE tenant_id = $1
        ORDER BY created_at DESC
    `

	rows, err := s.exec.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*IntegrationConfig
	for rows.Next() {
		config, err := scanIntegrationConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, rows.Err()
}

func (s *pgIntegrationStore) Count(ctx context.Context, tenantID string) (int64, error) {
	const q = `
		SELECT COUNT(*) FROM integration_config
		WHERE tenant_id = $1
	`

	var count int64
	err := s.exec.QueryRow(ctx, q, tenantID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ============================================================================
// METADADOS DO KUBEX (StoreType Implementation)
// ============================================================================

func (s *pgIntegrationStore) GetType() (reflect.Type, string, error) {
	return reflect.TypeFor[IntegrationConfig](), "pg_integration_store", nil
}

func (s *pgIntegrationStore) GetName() string {
	return "pg_integration_store"
}

func (s *pgIntegrationStore) Validate() error {
	if s.exec == nil {
		return gl.Errorf("PGExecutor is nil")
	}
	if s.crypto == nil {
		return gl.Errorf("CryptoService is nil")
	}
	if len(s.dbKey) == 0 {
		return gl.Errorf("Database Master Key is empty")
	}
	return nil
}

func (s *pgIntegrationStore) Close() error {
	return nil
}

// ============================================================================
// HELPERS DE SCAN
// ============================================================================

func scanIntegrationConfig(scanner interface{ Scan(dest ...any) error }) (*IntegrationConfig, error) {
	var c IntegrationConfig
	err := scanner.Scan(
		&c.ID,
		&c.TenantID,
		&c.PartnerID,
		&c.Type,
		&c.Name,
		&c.Settings,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func scanSyncJob(scanner interface{ Scan(dest ...any) error }) (*SyncJob, error) {
	var j SyncJob
	err := scanner.Scan(
		&j.ID,
		&j.TenantID,
		&j.ConfigID,
		&j.TaskName,
		&j.CronExpression,
		&j.IsActive,
		&j.LastSyncAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}
