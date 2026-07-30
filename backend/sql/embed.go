package dbschema

import _ "embed"

// InitialPostgresSchema is the canonical schema used by both startup migration
// and manual database initialization.
//
//go:embed 001_init_postgres.sql
var InitialPostgresSchema string

// AddUserContactsPostgres adds email and phone to databases created before
// those fields were part of the initial schema.
//
//go:embed 002_add_user_contacts_postgres.sql
var AddUserContactsPostgres string
