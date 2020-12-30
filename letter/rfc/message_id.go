package rfc

import (
	"fmt"

	"github.com/google/uuid"
)

type uuidGenerator struct {
	domain string
}

// UUIDGenerator returns a Message-ID factory using UUIDs. If domain is an
// empty string, it is set to "localhost". The generated IDs have the following
// format: <UUID@DOMAIN>
func UUIDGenerator(domain string) MessageIDFactory {
	if domain == "" {
		domain = "localhost"
	}
	return uuidGenerator{domain: domain}
}

func (gen uuidGenerator) GenerateID(Mail) string {
	return fmt.Sprintf("<%s@%s>", uuid.New().String(), gen.domain)
}
