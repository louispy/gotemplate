package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/louispy/gotemplate/internal/services"
)

type API struct {
	userService services.UserService
	// scaffold:services
	router *mux.Router
}

type Opts struct {
	UserService services.UserService
	// scaffold:opts
}

func NewAPI(o Opts) *API {
	r := mux.Router{}
	return &API{
		userService: o.UserService,
		// scaffold:assign
		router: &r,
	}
}

func (a *API) GetRouter() *mux.Router {
	return a.router
}

func (a *API) Register() {
	a.router.Methods(http.MethodGet).Path("/health").HandlerFunc(a.Health)

	a.router.Methods(http.MethodPost).Path("/users").HandlerFunc(a.CreateUser)
	a.router.Methods(http.MethodGet).Path("/users").HandlerFunc(a.ListUsers)
	a.router.Methods(http.MethodGet).Path("/users/{id}").HandlerFunc(a.GetUser)
	a.router.Methods(http.MethodPut).Path("/users/{id}").HandlerFunc(a.UpdateUser)
	a.router.Methods(http.MethodDelete).Path("/users/{id}").HandlerFunc(a.DeleteUser)
	// scaffold:routes
}
