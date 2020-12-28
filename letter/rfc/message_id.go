package rfc

import (
	"fmt"

	"github.com/google/uuid"
)

type uuidGenerator struct {
	domain string
}

// UUIDGenerator returns a Message-ID generator using UUIDs. If domain is an
// empty string, GenerateID() just returns a UUID. Otherwise the generated IDs
// have the following format: <UUID@DOMAIN>
func UUIDGenerator(domain string) IDGenerator {
	return uuidGenerator{
		domain: domain,
	}
}

func (gen uuidGenerator) GenerateID(Mail) string {
	id := uuid.New()
	if gen.domain == "" {
		return id.String()
	}
	return fmt.Sprintf("<%s@%s>", id.String(), gen.domain)
}
