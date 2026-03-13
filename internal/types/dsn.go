package types

import (
	"os"
	"regexp"
	"strings"

	"github.com/kubex-ecosystem/domus/internal/module/kbx"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// DSN is an interface for a connection string.
type DSN interface {
	// String returns the DSN string.
	String() string

	// RedactDSN removes sensitive information (passwords) from a DSN string.
	// Useful for safe logging of connection strings.
	//
	// Example:
	//
	//	postgres://user:secretpass@localhost:5432/db
	//	-> postgres://user:***@localhost:5432/db
	Redacted() string

	// GetOption returns the value of the option with the given key.
	GetOption(key string) (any, bool)
	// SetOption sets the value of the option with the given key.
	SetOption(key string, value any)
	// SetOptions sets the values of the options.
	SetOptions(options map[string]any)
	// GetOptions returns the options.
	GetOptions() map[string]any

	// Validate checks if the DSN is valid.
	Validate(opts ...string) error

	// Parse try to parse the DSN from a string.
	// It will replace the current DSN with the parsed DSN.
	// If the parse fails, it will return an error and the DSN will remain unchanged.
	Parse(raw string) error
}

// DSNImpl is a concrete implementation of the DSN interface.
// All fields are private to ensure encapsulation and controlled access.
type DSNImpl struct {
	protocol string
	user     string
	pass     string
	host     string
	port     string
	dbName   string
	schema   string
	tls      bool
	options  map[string]any
}

// NewDSNFromDBConfig constructs a connection string from database configuration components.
// Supports: PostgreSQL, MongoDB, Redis, RabbitMQ.
func NewDSNFromDBConfig(dbConfig kbx.DBConfig) DSN {
	dbType := string(dbConfig.Protocol)
	return &DSNImpl{
		protocol: dbType,
		user:     dbConfig.User,
		pass:     dbConfig.Pass,
		host:     dbConfig.Host,
		port:     dbConfig.Port,
		dbName:   dbConfig.DBName,
		schema:   dbConfig.Schema,
		tls:      dbConfig.TLSEnabled,
		options:  kbxGet.ValOrType(dbConfig.Options, make(map[string]any)),
	}
}

// NewDSN creates a new DSN from the given parameters.
func NewDSN(protocol, user, pass, host, port, dbName, schema string, tls bool, options map[string]any) DSN {
	return &DSNImpl{
		protocol: protocol,
		user:     user,
		pass:     pass,
		host:     host,
		port:     port,
		dbName:   dbName,
		schema:   schema,
		tls:      tls,
		options:  kbxGet.ValOrType(options, make(map[string]any)),
	}
}

func (d *DSNImpl) String() string {
	var strBuilder strings.Builder
	strBuilder.WriteString(d.protocol)
	strBuilder.WriteString("://")
	strBuilder.WriteString(d.user)
	strBuilder.WriteString(":")
	strBuilder.WriteString(kbxGet.EnvOr(os.ExpandEnv(d.pass), d.pass))
	strBuilder.WriteString("@")
	strBuilder.WriteString(d.host)
	strBuilder.WriteString(":")
	strBuilder.WriteString(d.port)
	strBuilder.WriteString("/")
	strBuilder.WriteString(d.dbName)
	strBuilder.WriteString("?")

	if len(d.schema) > 0 {
		strBuilder.WriteString("schema=")
		strBuilder.WriteString(d.schema)
		strBuilder.WriteString("&")
	}

	if d.tls {
		strBuilder.WriteString("tls=true")
		strBuilder.WriteString("&")
	}

	for k, v := range d.options {
		strBuilder.WriteString(k)
		strBuilder.WriteString("=")
		strBuilder.WriteString(gl.Sprintf("%v", v))
		strBuilder.WriteString("&")
	}

	return strings.TrimSuffix(strBuilder.String(), "&")
}

func (d *DSNImpl) Redacted() string {
	// Regex: match protocol://username:password@host
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	return re.ReplaceAllString(d.String(), "://$1:***@")
}

func (d *DSNImpl) GetOption(key string) (any, bool) {
	val, ok := d.GetOptions()[key]
	return val, ok
}

func (d *DSNImpl) SetOption(key string, value any) {
	d.options = kbxGet.ValOrType(d.options, make(map[string]any))
	d.options[key] = value
}

func (d *DSNImpl) SetOptions(options map[string]any) {
	d.options = kbxGet.ValOrType(options, make(map[string]any))
}

func (d *DSNImpl) GetOptions() map[string]any {
	return kbxGet.ValOrType(d.options, make(map[string]any))
}

func (d *DSNImpl) Validate(opts ...string) error {
	if len(d.protocol) == 0 {
		return gl.Errorf("Protocol is empty")
	}
	if len(d.user) == 0 {
		return gl.Errorf("User is empty")
	}
	if len(d.host) == 0 {
		return gl.Errorf("Host is empty")
	}
	if len(d.port) == 0 {
		return gl.Errorf("Port is empty")
	}
	if len(d.dbName) == 0 {
		return gl.Errorf("Name is empty")
	}
	return nil
}

func (d *DSNImpl) Parse(raw string) error {
	dsn, err := parseDSN(raw)
	if err != nil {
		return err
	}
	d.protocol = dsn.protocol
	d.user = dsn.user
	d.pass = dsn.pass
	d.host = dsn.host
	d.port = dsn.port
	d.dbName = dsn.dbName
	d.schema = dsn.schema
	d.tls = dsn.tls
	d.options = dsn.options
	return nil
}

// ----------- Private Methods -----------

func parseDSN(raw string) (*DSNImpl, error) {
	// TODO: Implementar parse do DSN
	if len(raw) == 0 {
		return nil, gl.Errorf("DSN is empty")
	}

	return nil, gl.Errorf("Not implemented")
}
