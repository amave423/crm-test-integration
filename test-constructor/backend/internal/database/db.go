package database

import (
    "fmt"
    "log"
    "test-constructor/config"
    "test-constructor/internal/models"
    "test-constructor/migrations"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
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

    connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("database connection failed: ", err)
    }

    DB = connection

    err = DB.AutoMigrate(
        &models.User{},
        &models.Test{},
        &models.Question{},
        &models.Answer{},
        &models.Role{},
        &models.Attempt{},
        &models.EventConfig{},
        &models.ExtraThreshold{},
    )
    if err != nil {
        log.Fatal("database migration failed: ", err)
    }

    if err := migrations.SeedRoles(DB); err != nil {
        log.Fatal("failed to seed roles: ", err)
    }

    if err := migrations.SeedAdmin(DB); err != nil {
        log.Fatal("failed to seed admin: ", err)
    }

    log.Println("database connected")
}
