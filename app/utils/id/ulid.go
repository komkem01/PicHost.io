package id

import (
	"github.com/oklog/ulid/v2"
)

func NewULID() ulid.ULID {
	return ulid.Make()
}
