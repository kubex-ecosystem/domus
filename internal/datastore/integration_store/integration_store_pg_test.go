package integrationstore

import (
	"encoding/json"
	"strings"
	"testing"
)

// newCryptoStore monta um pgIntegrationStore SÓ com o necessário para exercitar
// os helpers de cifragem (sem banco). A chave-mestra pode ter QUALQUER tamanho —
// encKey() a deriva para 32 bytes via SHA-256. Aqui usamos uma string comum, de
// propósito: prova que não dependemos mais do formato de chave da lib ethyr.
func newCryptoStore(t *testing.T) *pgIntegrationStore {
	t.Helper()
	return &pgIntegrationStore{dbKey: []byte("qualquer-chave-mestra-de-dev")}
}

// asMap desserializa um RawMessage para inspeção.
func asMap(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestEncryptAninhado(t *testing.T) {
	s := newCryptoStore(t)

	// Shape real de uma conexão de marketplace: o segredo está DENTRO de
	// credentials (objeto aninhado). É exatamente o que ficava em texto puro.
	settings := json.RawMessage(`{
		"kind": "marketplace",
		"platform": "ifood",
		"webhook_secret": "wh-super-secreto",
		"credentials": {
			"client_id": "cid-publico",
			"client_secret": "cs-super-secreto",
			"merchant_id": "m-123"
		}
	}`)

	enc, err := s.encryptSensitiveFields(settings)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	encMap := asMap(t, enc)

	// 1. webhook_secret (topo) — selado.
	if ws, _ := encMap["webhook_secret"].(string); !strings.HasPrefix(ws, encPrefixV2) {
		t.Errorf("webhook_secret deveria estar selado, veio %q", ws)
	}

	creds, ok := encMap["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("credentials deveria continuar sendo objeto, veio %T", encMap["credentials"])
	}

	// 2. O FIX: client_secret ANINHADO agora está selado (antes: texto puro).
	if cs, _ := creds["client_secret"].(string); !strings.HasPrefix(cs, encPrefixV2) {
		t.Errorf("client_secret aninhado deveria estar selado, veio %q", cs)
	}
	// Garantia dura: o texto puro NÃO pode aparecer em lugar nenhum do JSON cifrado.
	if strings.Contains(string(enc), "cs-super-secreto") || strings.Contains(string(enc), "wh-super-secreto") {
		t.Error("segredo em texto puro vazou no JSONB cifrado")
	}

	// 3. Campos não-sensíveis permanecem em claro.
	if creds["client_id"] != "cid-publico" || creds["merchant_id"] != "m-123" {
		t.Errorf("campos não-sensíveis não deveriam ser cifrados: %+v", creds)
	}
	if encMap["platform"] != "ifood" {
		t.Errorf("platform não deveria ser cifrado, veio %v", encMap["platform"])
	}
}

func TestRoundTripDeterministico(t *testing.T) {
	s := newCryptoStore(t)

	// Segredos que ANTES quebravam por serem "base64-like" (16 chars c/ hífen) ou
	// por cair no bug do TrimSpace. Agora o round-trip é 100% determinístico.
	secrets := []string{
		"wh-super-secreto",
		"cs-super-secreto",
		"segredo-x",
		"AKIA1234567890ABCDEF",         // parece credencial AWS (base64-like)
		"ghp_aBcDeFgHiJkLmNoPqRsTuVwX", // parece token (base64-like)
		"senha!@#$%",
		"x",
	}
	for _, sec := range secrets {
		settings := json.RawMessage(`{"credentials":{"client_secret":` + mustJSON(sec) + `}}`)
		enc, err := s.encryptSensitiveFields(settings)
		if err != nil {
			t.Fatalf("encrypt %q: %v", sec, err)
		}
		if strings.Contains(string(enc), sec) && sec != "x" {
			// (sec "x" é curto demais para afirmar não-vazamento por substring)
			t.Errorf("segredo %q vazou em texto puro", sec)
		}
		dec, err := s.decryptSensitiveFields(enc)
		if err != nil {
			t.Fatalf("decrypt %q: %v", sec, err)
		}
		got := asMap(t, dec)["credentials"].(map[string]any)["client_secret"]
		if got != sec {
			t.Errorf("round-trip falhou: quero %q, veio %q", sec, got)
		}
	}
}

// TestRoundTripVolume garante que não há corrupção intermitente (o bug do
// TrimSpace da lib ethyr causava ~5% de falha; aqui deve ser 0%).
func TestRoundTripVolume(t *testing.T) {
	s := newCryptoStore(t)
	const N = 1000
	for i := 0; i < N; i++ {
		settings := json.RawMessage(`{"credentials":{"client_secret":"wh-super-secreto"}}`)
		enc, err := s.encryptSensitiveFields(settings)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		dec, err := s.decryptSensitiveFields(enc)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}
		if got := asMap(t, dec)["credentials"].(map[string]any)["client_secret"]; got != "wh-super-secreto" {
			t.Fatalf("corrupção na iteração %d: veio %q", i, got)
		}
	}
}

func TestEncryptIdempotente(t *testing.T) {
	// Cifrar duas vezes não deve re-selar (o prefixo enc_v2: barra a dupla selagem).
	s := newCryptoStore(t)
	settings := json.RawMessage(`{"credentials":{"client_secret":"x"}}`)

	enc1, err := s.encryptSensitiveFields(settings)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := s.encryptSensitiveFields(enc1)
	if err != nil {
		t.Fatal(err)
	}
	c1 := asMap(t, enc1)["credentials"].(map[string]any)["client_secret"]
	c2 := asMap(t, enc2)["credentials"].(map[string]any)["client_secret"]
	if c1 != c2 {
		t.Errorf("dupla selagem alterou o valor: %v != %v", c1, c2)
	}
}

func TestLegadoV1EhTextoPuro(t *testing.T) {
	// O esquema antigo (enc_v1:) nunca cifrou — selava o texto puro. Ao ler, o
	// valor após o prefixo deve ser devolvido como está (migração na escrita).
	s := newCryptoStore(t)
	legacy := json.RawMessage(`{"credentials":{"client_secret":"enc_v1:era-texto-puro"}}`)

	dec, err := s.decryptSensitiveFields(legacy)
	if err != nil {
		t.Fatalf("decrypt legado: %v", err)
	}
	if got := asMap(t, dec)["credentials"].(map[string]any)["client_secret"]; got != "era-texto-puro" {
		t.Errorf("legado v1 deveria virar texto puro, veio %q", got)
	}
}

func TestDecryptRetrocompativelTextoPuro(t *testing.T) {
	// Dado sem selo (texto puro) passa intacto na leitura.
	s := newCryptoStore(t)
	legacy := json.RawMessage(`{"credentials":{"client_secret":"texto-puro-antigo"}}`)

	dec, err := s.decryptSensitiveFields(legacy)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got := asMap(t, dec)["credentials"].(map[string]any)["client_secret"]; got != "texto-puro-antigo" {
		t.Errorf("texto puro deveria passar intacto, veio %q", got)
	}
}

func TestEmptyRawIntacto(t *testing.T) {
	s := newCryptoStore(t)
	for _, raw := range []json.RawMessage{nil, {}} {
		enc, err := s.encryptSensitiveFields(raw)
		if err != nil {
			t.Errorf("encrypt de vazio não deveria falhar: %v", err)
		}
		if len(enc) != 0 {
			t.Errorf("encrypt de vazio deveria devolver vazio, veio %q", enc)
		}
	}
}

func mustJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
