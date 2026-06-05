package sqlsafe

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		// ── válidos ──────────────────────────────────────────────────────────────
		{name: "select simples", sql: "SELECT id FROM users", wantErr: false},
		{name: "select com where", sql: "SELECT * FROM orders WHERE id = 1", wantErr: false},
		{name: "select com semicolon final", sql: "SELECT 1;", wantErr: false},
		{name: "select maiúsculo", sql: "SELECT VLRNOTA FROM TGFCAB", wantErr: false},
		{name: "select com join", sql: "SELECT a.id, b.name FROM a JOIN b ON a.id = b.a_id", wantErr: false},
		{name: "select com subquery", sql: "SELECT * FROM (SELECT id FROM x) t", wantErr: false},
		// colunas cujo nome contém keyword — não devem ser rejeitadas
		{name: "coluna last_updated", sql: "SELECT last_updated FROM audit_log", wantErr: false},
		{name: "coluna total_created", sql: "SELECT total_created FROM stats", wantErr: false},

		// ── inválidos: DML ────────────────────────────────────────────────────────
		{name: "insert", sql: "INSERT INTO users (name) VALUES ('x')", wantErr: true},
		{name: "update", sql: "UPDATE users SET name='x' WHERE id=1", wantErr: true},
		{name: "delete", sql: "DELETE FROM users WHERE id=1", wantErr: true},
		{name: "insert minúsculo", sql: "insert into users (name) values ('x')", wantErr: true},
		{name: "update minúsculo", sql: "update users set name='x'", wantErr: true},

		// ── inválidos: DDL ────────────────────────────────────────────────────────
		{name: "drop", sql: "DROP TABLE users", wantErr: true},
		{name: "alter", sql: "ALTER TABLE users ADD COLUMN foo TEXT", wantErr: true},
		{name: "truncate", sql: "TRUNCATE TABLE users", wantErr: true},
		{name: "create", sql: "CREATE TABLE x (id INT)", wantErr: true},
		{name: "drop minúsculo", sql: "drop table users", wantErr: true},

		// ── inválidos: multi-statement ────────────────────────────────────────────
		{name: "dois selects", sql: "SELECT 1; SELECT 2", wantErr: true},
		{name: "select + delete", sql: "SELECT 1; DELETE FROM users", wantErr: true},
		{name: "três statements", sql: "SELECT 1; SELECT 2; SELECT 3", wantErr: true},

		// ── inválidos: vazio ──────────────────────────────────────────────────────
		{name: "string vazia", sql: "", wantErr: true},
		{name: "só espaços", sql: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr = %v", tt.sql, err, tt.wantErr)
			}
		})
	}
}
