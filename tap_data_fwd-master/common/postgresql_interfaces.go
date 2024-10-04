package common

import (
	"gorm.io/gorm"
)

type DBConfigLoader interface {
	LoadPostgreSQLConfig() int
}

type DBConnector interface {
	GetDatabaseConnection() int
}

type DBAccessor interface {
	GetDB() *gorm.DB
}

type PostgreSQLConfig struct {
	Username string
	Password string
	DBName   string
	Host     string
	Port     int
	SSLMode  string
}
