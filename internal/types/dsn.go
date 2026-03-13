package types

import (
	"reflect"
	"regexp"

	"github.com/kubex-ecosystem/domus/internal/module/kbx"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

// DSN is an interface for a connection string.
type DSN[T Driver] interface {
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

	// Flavor returns the flavor of the DSN.
	Driver() reflect.Type

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
type DSNImpl[T Driver] struct {
	protocol string
	user     string
	pass     string
	host     string
	port     string
	name     string
	options  map[string]any
}

// NewDSNFromDBConfig constructs a connection string from database configuration components.
// Supports: PostgreSQL, MongoDB, Redis, RabbitMQ.
func NewDSNFromDBConfig[T Driver](dbConfig kbx.DBConfig) DSN[T] {
	dbType := string(dbConfig.Protocol)
	return &DSNImpl[T]{
		protocol: dbType,
		user:     dbConfig.User,
		pass:     dbConfig.Pass,
		host:     dbConfig.Host,
		port:     dbConfig.Port,
		name:     dbConfig.Name,
		options:  kbxGet.ValOrType(dbConfig.Options, make(map[string]any)),
	}
}

// NewDSN creates a new DSN from the given parameters.
func NewDSN[T Driver](protocol, user, pass, host, port, name string, options map[string]any) DSN[T] {
	return &DSNImpl[T]{
		protocol: protocol,
		user:     user,
		pass:     pass,
		host:     host,
		port:     port,
		name:     name,
		options:  kbxGet.ValOrType(options, make(map[string]any)),
	}
}

func (d *DSNImpl[T]) Driver() reflect.Type { return reflect.TypeFor[T]() }

func (d *DSNImpl[T]) String() string {
	return gl.Sprintf("%s://%s:%s@%s:%s/%s",
		d.protocol,
		d.user,
		kbxGet.EnvOr(d.pass, d.pass),
		d.host,
		d.port,
		d.name,
	)
}

func (d *DSNImpl[T]) Redacted() string {
	// Regex: match protocol://username:password@host
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	return re.ReplaceAllString(d.String(), "://$1:***@")
}

func (d *DSNImpl[T]) GetOption(key string) (any, bool) {
	val, ok := d.GetOptions()[key]
	return val, ok
}

func (d *DSNImpl[T]) SetOption(key string, value any) {
	d.options = kbxGet.ValOrType(d.options, make(map[string]any))
	d.options[key] = value
}

func (d *DSNImpl[T]) SetOptions(options map[string]any) {
	d.options = kbxGet.ValOrType(options, make(map[string]any))
}

func (d *DSNImpl[T]) GetOptions() map[string]any {
	return kbxGet.ValOrType(d.options, make(map[string]any))
}

func (d *DSNImpl[T]) Validate(opts ...string) error {
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
	if len(d.name) == 0 {
		return gl.Errorf("Name is empty")
	}
	return nil
}

func (d *DSNImpl[T]) Parse(raw string) error {
	dsn, err := parseDSN(raw)
	if err != nil {
		return err
	}
	d.protocol = dsn.protocol
	d.user = dsn.user
	d.pass = dsn.pass
	d.host = dsn.host
	d.port = dsn.port
	d.name = dsn.name
	d.options = dsn.options
	return nil
}

// ----------- Private Methods -----------

func parseDSN(raw string) (*DSNImpl[Driver], error) {
	// TODO: Implementar parse do DSN
	if len(raw) == 0 {
		return nil, gl.Errorf("DSN is empty")
	}

	return nil, gl.Errorf("Not implemented")
}
