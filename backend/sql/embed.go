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

// AddPaymentDetailsPostgres adds provider identifiers and QR payment metadata.
//
//go:embed 005_add_payment_details_postgres.sql
var AddPaymentDetailsPostgres string

// AddLicenseCodesPostgres adds one-time tool redemption codes.
//
//go:embed 006_add_license_codes_postgres.sql
var AddLicenseCodesPostgres string

// AddGameTablesPostgres adds game system tables for the box-opening RPG game.
//
//go:embed 007_add_game_tables_postgres.sql
var AddGameTablesPostgres string
