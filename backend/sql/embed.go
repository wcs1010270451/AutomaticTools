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

// AddAdminsPostgres adds administrator accounts used by management APIs.
//
//go:embed 003_add_admins_postgres.sql
var AddAdminsPostgres string

// AddEmailVerificationCodesPostgres adds one-time email verification records.
//
//go:embed 004_add_email_verification_codes_postgres.sql
var AddEmailVerificationCodesPostgres string
