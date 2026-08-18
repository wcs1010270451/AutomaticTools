package store

import (
	"context"
	"errors"
	"strings"
	"time"

	dbschema "automatictools/backend/sql"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Open(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	migrations := []string{
		dbschema.InitialPostgresSchema,
		dbschema.AddUserContactsPostgres,
		dbschema.AddAdminsPostgres,
		dbschema.AddEmailVerificationCodesPostgres,
		dbschema.AddPaymentDetailsPostgres,
		dbschema.AddLicenseCodesPostgres,
		dbschema.AddGameTablesPostgres,
	}
	for _, migration := range migrations {
		if err := db.Exec(migration).Error; err != nil {
			return err
		}
	}
	return nil
}

// EnsureDefaultAdmin creates the initial administrator if it does not exist.
// Existing accounts are never overwritten, so changing the bootstrap password
// after the first startup does not reset an administrator's password.
func EnsureDefaultAdmin(db *gorm.DB, username string, password string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return errors.New("default admin username cannot be empty")
	}
	if len(password) < 6 {
		return errors.New("default admin password must contain at least 6 characters")
	}

	var existing Admin
	err := db.Where("username = ?", username).Take(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	admin := Admin{
		Username:     username,
		PasswordHash: string(passwordHash),
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&admin).Error
}
