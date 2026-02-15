package types

import gl "github.com/kubex-ecosystem/logz"

type DSN interface {
	String() string
	GetFlavor() string
	Redact() string
	IsEmpty() bool
	GetOption(key string) string
	SetOption(key string, value string)
	ClearOptions()
	GetOptions() map[string]string
	HasOptions() bool
	Validate(opts ...string) error
}

// DSNImpl is a concrete implementation of the DSN interface.
// All fields are private to ensure encapsulation and controlled access.
type DSNImpl struct {
	raw      *string
	protocol string
	user     string
	pass     string
	host     string
	port     string
	name     string
	options  map[string]string
}

func NewDSN(raw *string) DSN {
	return &DSNImpl{
		raw: raw,
	}
}

func (d *DSNImpl) String() string {
	if d == nil {
		return ""
	}
	return gl.Sprintf("%s://%s:%s@%s:%s/%s", d.protocol, d.user, d.pass, d.host, d.port, d.name)
}

func (d *DSNImpl) GetFlavor() string {
	if d == nil {
		return ""
	}
	switch d.protocol {
	case "postgres", "postgresql":
		return "postgres"
	case "mongodb", "mongo":
		return "mongo"
	case "redis":
		return "redis"
	case "rabbitmq", "amqp":
		return "rabbitmq"
	default:
		return ""
	}
}

func (d *DSNImpl) Redact() string {
	if d == nil {
		return ""
	}
	return gl.Sprintf("%s://%s:***@%s:%s/%s", d.protocol, d.user, d.host, d.port, d.name)
}

func (d *DSNImpl) IsEmpty() bool {
	if d == nil {
		return true
	}
	return d.user == "" && d.pass == "" && d.host == "" && d.port == "" && d.name == ""
}

func (d *DSNImpl) GetOption(key string) string {
	if d == nil || d.options == nil {
		return ""
	}
	return d.options[key]
}

func (d *DSNImpl) SetOption(key string, value string) {
	if d == nil {
		return
	}
	if d.options == nil {
		d.options = make(map[string]string)
	}
	d.options[key] = value
}

func (d *DSNImpl) ClearOptions() {
	if d == nil {
		return
	}
	d.options = nil
}

func (d *DSNImpl) GetOptions() map[string]string {
	if d == nil {
		return nil
	}
	return d.options
}

func (d *DSNImpl) HasOptions() bool {
	if d == nil {
		return false
	}
	return len(d.options) > 0
}

func (d *DSNImpl) Validate(opts ...string) error {
	s := d // Preparados para tratativa com o opts

	if s == nil {
		return gl.Errorf("DSN is nil")
	}
	if s.protocol == "" {
		return gl.Errorf("DSN protocol is empty")
	}
	if s.user == "" {
		return gl.Errorf("DSN user is empty")
	}
	if s.host == "" {
		return gl.Errorf("DSN host is empty")
	}
	if s.port == "" {
		return gl.Errorf("DSN port is empty")
	}
	if s.name == "" {
		return gl.Errorf("DSN name is empty")
	}
	return nil
}
