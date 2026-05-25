package datasource

import (
	"strings"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// context_ is the standard datasource.Context implementation. It stores
// the structured connection parameters and pre-expands the placeholders
// :username, :password, :server and :service in the URL at construction
// time so drivers receive a ready-to-open DSN.
type context_ struct {
	url      string
	user     string
	password string
	server   string
	service  string
}

// NewContext builds a Context from the structured connection parameters.
//
// Every argument is required. The url MAY contain the placeholders
// :username, :password, :server and :service; NewContext substitutes
// them with the matching arguments so drivers obtain a fully formed DSN
// through Url().
func NewContext(url string, user string, password string, server string, service string) Context {
	cassert.NotEmpty(url, "url is empty")
	cassert.NotEmpty(user, "user is empty")
	cassert.NotEmpty(password, "password is empty")
	cassert.NotEmpty(server, "server is empty")
	cassert.NotEmpty(service, "service is empty")

	url = strings.Replace(url, ":username", user, 1)
	url = strings.Replace(url, ":password", password, 1)
	url = strings.Replace(url, ":server", server, 1)
	url = strings.Replace(url, ":service", service, 1)

	return &context_{
		url:      url,
		user:     user,
		password: password,
		server:   server,
		service:  service,
	}
}

// Url returns the DSN template after placeholder substitution.
func (c *context_) Url() string {
	cassert.NotNil(c, "context is nil")

	return c.url
}

// User returns the authentication user.
func (c *context_) User() string {
	cassert.NotNil(c, "context is nil")

	return c.user
}

// Password returns the authentication password.
func (c *context_) Password() string {
	cassert.NotNil(c, "context is nil")

	return c.password
}

// Server returns the server endpoint.
func (c *context_) Server() string {
	cassert.NotNil(c, "context is nil")

	return c.server
}

// Service returns the logical service/database identifier.
func (c *context_) Service() string {
	cassert.NotNil(c, "context is nil")

	return c.service
}
