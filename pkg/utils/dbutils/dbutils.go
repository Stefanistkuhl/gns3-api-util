package dbutils

import (
	"database/sql"
	"time"
)

func NullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func NullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func NullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func StringPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

func Int64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func BoolPtr(n sql.NullBool) *bool {
	if !n.Valid {
		return nil
	}
	return &n.Bool
}

func TimePtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	return &n.Time
}
