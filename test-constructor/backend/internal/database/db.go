package database

import (
	"fmt"
	"log"
	"test-constructor/config"
	"test-constructor/internal/domain"
	"test-constructor/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	cfg := config.Load()
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
			cfg.DBPort,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}

	err = db.AutoMigrate(
		&domain.User{},
		&domain.Test{},
		&domain.Question{},
		&domain.Answer{},
		&domain.Role{},
		&domain.EventConfig{},
		&domain.ExtraThreshold{},
		&domain.Attempt{},
		&domain.UserEvent{},
	)
	if err != nil {
		log.Fatal("database migration failed: ", err)
	}

	fixLegacyConstraints(db)

	if err := migrations.SeedRoles(db); err != nil {
		log.Fatal("failed to seed roles: ", err)
	}

	if err := migrations.SeedAdmin(db); err != nil {
		log.Fatal("failed to seed admin: ", err)
	}

	log.Println("database connected")
	return db
}

func fixLegacyConstraints(db *gorm.DB) {
	if err := db.Exec(`ALTER TABLE event_configs DROP CONSTRAINT IF EXISTS fk_attempts_event_config`).Error; err != nil {
		log.Println("failed to drop legacy event config constraint: ", err)
	}

	if err := db.Exec(`ALTER TABLE attempts DROP CONSTRAINT IF EXISTS fk_attempts_event_config`).Error; err != nil {
		log.Println("failed to drop attempts event config constraint: ", err)
	}

	if err := db.Exec(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_attempts_event_config'
          AND conrelid = 'attempts'::regclass
    ) THEN
        ALTER TABLE attempts
        ADD CONSTRAINT fk_attempts_event_config
        FOREIGN KEY (config_id)
        REFERENCES event_configs(config_id)
        ON DELETE CASCADE;
    END IF;
END
$$;
`).Error; err != nil {
		log.Println("failed to create attempts event config constraint: ", err)
	}
}
