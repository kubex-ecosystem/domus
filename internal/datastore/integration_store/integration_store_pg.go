package integrationstore

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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
	"golang.org/x/crypto/chacha20poly1305"
)

// pgIntegrationStore implementa IntegrationStore usando PGExecutor. Os campos
// sensíveis do JSONB são selados com XChaCha20-Poly1305 (ver sealSecret) usando
// uma chave derivada da master key (dbKey).
type pgIntegrationStore struct {
	exec  execution.PGExecutor
	dbKey []byte // Master key injetada na inicialização do store
}

// NewPGIntegrationStore cria uma instância do store com criptografia embutida.
func NewPGIntegrationStore(exec execution.PGExecutor, masterKey []byte) IntegrationStore {
	return &pgIntegrationStore{
		exec:  exec,
		dbKey: masterKey,
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

// ListActiveConfigs lista as configs ativas de todos os tenants.
//
// NÃO decifra o Settings de propósito: o consumidor típico é um processo de
// plano de fundo que só precisa saber quais conexões existem. Decifrar a lista
// inteira colocaria as credenciais de todos os tenants em memória de uma vez,
// sem necessidade. Quem precisa dos segredos chama GetConfigByID.
func (s *pgIntegrationStore) ListActiveConfigs(ctx context.Context) ([]IntegrationConfig, error) {
	const q = `
		SELECT id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
		FROM integration_config
		WHERE is_active = true
		ORDER BY created_at
	`

	rows, err := s.exec.Query(ctx, q)
	if err != nil {
		return nil, gl.Errorf("failed to list active integration configs: %v", err)
	}
	defer rows.Close()

	var configs []IntegrationConfig
	for rows.Next() {
		config, err := scanIntegrationConfig(rows)
		if err != nil {
			return nil, gl.Errorf("failed to scan integration config: %v", err)
		}
		configs = append(configs, *config)
	}
	return configs, rows.Err()
}

// UpdateConfig atualiza a configuração blindando os dados sensíveis do JSONB.
// Escopado por tenant: a cláusula WHERE exige tenant_id, impedindo que um tenant
// sobrescreva a config de outro mesmo conhecendo o ID.
func (s *pgIntegrationStore) UpdateConfig(ctx context.Context, input *IntegrationConfig) (*IntegrationConfig, error) {
	if input == nil {
		return nil, gl.Errorf("update input is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.ID) == "" {
		return nil, gl.Errorf("id is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.TenantID) == "" {
		return nil, gl.Errorf("tenant id is required: %v", t.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, gl.Errorf("name is required: %v", t.ErrInvalidInput)
	}

	// Sela os segredos antes de tocar o banco — mesmo caminho do CreateConfig.
	secureSettings, err := s.encryptSensitiveFields(input.Settings)
	if err != nil {
		return nil, gl.Errorf("failed to encrypt settings: %v", err)
	}

	const q = `
		UPDATE integration_config
		SET partner_id = $3, type = $4, name = $5, settings = $6, is_active = $7, updated_at = $8
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, partner_id, type, name, settings, is_active, created_at, updated_at
	`

	row := s.exec.QueryRow(ctx, q,
		input.ID,
		input.TenantID,
		input.PartnerID,
		input.Type,
		input.Name,
		secureSettings,
		input.IsActive,
		time.Now().UTC(),
	)

	config, err := scanIntegrationConfig(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // inexistente ou de outro tenant
		}
		return nil, gl.Errorf("failed to update integration config: %v", err)
	}

	config.Settings, _ = s.decryptSensitiveFields(config.Settings)
	return config, nil
}

// DeleteConfig remove a configuração do tenant informado.
func (s *pgIntegrationStore) DeleteConfig(ctx context.Context, tenantID, id string) error {
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(id) == "" {
		return gl.Errorf("tenant id and id are required: %v", t.ErrInvalidInput)
	}

	const q = `DELETE FROM integration_config WHERE id = $1 AND tenant_id = $2`

	res, err := s.exec.Exec(ctx, q, id, tenantID)
	if err != nil {
		return gl.Errorf("failed to delete integration config: %v", err)
	}
	if res.RowsAffected() == 0 {
		return gl.Errorf("integration config %s not found for tenant", id)
	}
	return nil
}

// ============================================================================
// HELPERS DE CRIPTOGRAFIA DINÂMICA
// ============================================================================

// encryptSensitiveFields varre o JSON procurando segredos e os criptografa.
//
// A varredura é RECURSIVA: chaves sensíveis são seladas em qualquer profundidade,
// não só no nível superior. Isto é o que protege segredos aninhados como
// `credentials.client_secret` — antes a varredura era de 1 nível só, e um objeto
// aninhado (como `credentials`) ficava inteiro em texto puro no banco.
func (s *pgIntegrationStore) encryptSensitiveFields(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if err := s.encryptNode(data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// encryptNode cifra, in-place, os valores string sob chaves sensíveis, descendo
// em objetos e arrays. A decisão de cifrar é por CHAVE (isSensitiveKey); por isso
// strings soltas dentro de arrays (sem chave associada) não são seladas.
func (s *pgIntegrationStore) encryptNode(node map[string]any) error {
	for key, value := range node {
		switch v := value.(type) {
		case string:
			if !isSensitiveKey(key) || v == "" || strings.HasPrefix(v, encPrefixV2) {
				continue // não sensível, vazio, ou já selado (evita dupla encriptação)
			}
			sealed, err := s.sealSecret(v)
			if err != nil {
				return gl.Errorf("encryption failed for key %s: %w", key, err)
			}
			node[key] = sealed
		case map[string]any:
			if err := s.encryptNode(v); err != nil {
				return err
			}
		case []any:
			if err := s.encryptArray(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *pgIntegrationStore) encryptArray(arr []any) error {
	for _, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			if err := s.encryptNode(v); err != nil {
				return err
			}
		case []any:
			if err := s.encryptArray(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// decryptSensitiveFields varre o JSON e restaura os valores originais.
//
// Simétrico ao encrypt: recursivo, decidindo pelo selo (`enc_v2:` novo, `enc_v1:`
// legado). Retrocompatível — valores sem selo, ou selados pelo esquema antigo que
// nunca cifrava, são devolvidos como texto e passam a ser cifrados de verdade na
// próxima escrita.
func (s *pgIntegrationStore) decryptSensitiveFields(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if err := s.decryptNode(data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (s *pgIntegrationStore) decryptNode(node map[string]any) error {
	for key, value := range node {
		switch v := value.(type) {
		case string:
			switch {
			case strings.HasPrefix(v, encPrefixV2):
				plain, err := s.openSecret(v)
				if err != nil {
					return gl.Errorf("decryption failed for key %s: %w", key, err)
				}
				node[key] = plain
			case strings.HasPrefix(v, encPrefixV1):
				// Formato legado: o esquema antigo NUNCA cifrou de fato (selava o
				// texto puro com o prefixo por causa de um bug no uso do
				// CryptoService). O valor após o prefixo é o texto puro — devolve-o.
				// Na próxima escrita ele passa a ser selado de verdade (enc_v2).
				node[key] = strings.TrimPrefix(v, encPrefixV1)
			}
			// sem prefixo → texto puro legado, passa intacto (será selado na escrita).
		case map[string]any:
			if err := s.decryptNode(v); err != nil {
				return err
			}
		case []any:
			if err := s.decryptArray(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *pgIntegrationStore) decryptArray(arr []any) error {
	for _, item := range arr {
		switch v := item.(type) {
		case map[string]any:
			if err := s.decryptNode(v); err != nil {
				return err
			}
		case []any:
			if err := s.decryptArray(v); err != nil {
				return err
			}
		}
	}
	return nil
}

// ============================================================================
// SELAGEM DE SEGREDOS (XChaCha20-Poly1305 direto)
// ============================================================================

const (
	// encPrefixV2 marca um valor selado pelo esquema atual: XChaCha20-Poly1305
	// sobre o texto UTF-8, transportado em base64 padrão (com padding).
	encPrefixV2 = "enc_v2:"
	// encPrefixV1 é o esquema legado, que na prática NÃO cifrava — ver a nota no
	// decryptNode. Mantido só para ler dados antigos.
	encPrefixV1 = "enc_v1:"
)

// sealSecret cifra um segredo e devolve "enc_v2:<base64(nonce||ciphertext)>".
//
// Usa a primitiva XChaCha20-Poly1305 diretamente (mesmo algoritmo do
// CryptoService do ethyr), em vez do Encrypt/Decrypt de alto nível daquela lib.
// Motivo: aquele transporte aplica bytes.TrimSpace sobre o ciphertext BINÁRIO e
// auto-decodifica inputs que "parecem base64", o que corrompe ~5% dos segredos e
// altera silenciosamente strings base64-like. Aqui o caminho é determinístico:
// base64.StdEncoding (com padding, sem TrimSpace) sobre nonce||ciphertext.
func (s *pgIntegrationStore) sealSecret(plaintext string) (string, error) {
	aead, err := chacha20poly1305.NewX(s.encKey())
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	// Seal prependa o nonce (1º arg) e anexa o ciphertext+tag.
	sealed := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefixV2 + base64.StdEncoding.EncodeToString(sealed), nil
}

// openSecret desfaz sealSecret.
func (s *pgIntegrationStore) openSecret(sealed string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(sealed, encPrefixV2))
	if err != nil {
		return "", fmt.Errorf("invalid sealed payload: %w", err)
	}
	aead, err := chacha20poly1305.NewX(s.encKey())
	if err != nil {
		return "", err
	}
	if len(raw) < aead.NonceSize() {
		return "", fmt.Errorf("sealed payload too short")
	}
	nonce, ciphertext := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}

// encKey deriva uma chave de 32 bytes (KeySize do XChaCha) a partir da mestra,
// via SHA-256. Assim qualquer mKey (de qualquer tamanho, inclusive o fallback de
// dev do gnyx) produz uma chave válida e determinística — sem depender de a mKey
// já ter exatamente 32 bytes nem passar pela validação problemática do ethyr.
func (s *pgIntegrationStore) encKey() []byte {
	sum := sha256.Sum256(s.dbKey)
	return sum[:]
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

// Update delega para UpdateConfig. A versão anterior gravava config.Settings
// CRU no JSONB — ou seja, persistia credenciais em texto puro, furando o cofre.
func (s *pgIntegrationStore) Update(ctx context.Context, config *IntegrationConfig) error {
	if config == nil {
		return gl.Errorf("update input is required: %v", t.ErrInvalidInput)
	}
	updated, err := s.UpdateConfig(ctx, config)
	if err != nil {
		return err
	}
	if updated == nil {
		return gl.Errorf("integration config %s not found for tenant", config.ID)
	}
	*config = *updated
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
