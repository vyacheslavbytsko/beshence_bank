package misc

import "github.com/google/uuid"

func ParseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func SameOptionalUUID(a, b *uuid.UUID) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func OptionalUUIDToString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}

	return new(value.String())
}
