package dbrepo

import (
	"database/sql"

	"github.com/atuprosper/booking-project/internal/config"
	"github.com/atuprosper/booking-project/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(dbConnection *sql.DB, appConfig *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: appConfig,
		DB:  dbConnection,
	}
}

func NewTestRepo(appConfig *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: appConfig,
	}
}
