package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/louispy/template/internal/services"
)

type API struct {
	userService services.UserService
	router      *mux.Router
}

type Opts struct {
	UserService services.UserService
}

func NewAPI(o Opts) *API {
	r := mux.Router{}
	return &API{
		userService: o.UserService,
		router:      &r,
	}
}

func (a *API) GetRouter() *mux.Router {
	return a.router
}

func (a *API) Register() {
	a.router.Methods(http.MethodGet).Path("/hello").HandlerFunc(a.Hello)
	a.router.Methods(http.MethodGet).Path("/health").HandlerFunc(a.Health)

	a.router.Methods(http.MethodPost).Path("/users").HandlerFunc(a.CreateUser)
	a.router.Methods(http.MethodGet).Path("/users").HandlerFunc(a.ListUsers)
	a.router.Methods(http.MethodGet).Path("/users/{id}").HandlerFunc(a.GetUser)
	a.router.Methods(http.MethodPut).Path("/users/{id}").HandlerFunc(a.UpdateUser)
	a.router.Methods(http.MethodDelete).Path("/users/{id}").HandlerFunc(a.DeleteUser)
}
