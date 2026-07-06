package main

import (
	"context"

	"github.com/caarlos0/env/v11"
	"github.com/louispy/template/internal/api"
	"github.com/louispy/template/internal/database"
	"github.com/louispy/template/internal/domain/repositories"
	"github.com/louispy/template/internal/services"
)

type Container struct {
	API *api.API
}

type Config struct {
	DB database.Config `envPrefix:"DB_"`
}

func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic("cannot parse config: " + err.Error())
	}

	return &cfg
}

func NewContainer() *Container {
	cfg := NewConfig()

	db, err := database.New(context.Background(), cfg.DB)
	if err != nil {
		panic("database cannot be initialized: " + err.Error())
	}

	usersRepo := repositories.NewUsersRepository(repositories.UsersRepoOpts{DB: db})
	userService := services.NewUserService(services.UserServiceOpts{
		UsersRepo: usersRepo,
	})

	appAPI := api.NewAPI(api.Opts{
		UserService: userService,
	})
	appAPI.Register()

	return &Container{
		API: appAPI,
	}
}
