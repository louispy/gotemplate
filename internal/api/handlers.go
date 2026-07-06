package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/louispy/gotemplate/internal/services"
)

func (a API) Health(rw http.ResponseWriter, r *http.Request) {
	WriteJSONResponse(rw, http.StatusOK, map[string]string{"status": "ok"}, nil, nil)
}

func (a API) CreateUser(rw http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(rw, r.Body, 1<<20)
	req := CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}

	output, err := a.userService.Create(r.Context(), services.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		WriteJSONResponse(rw, statusForError(err), nil, nil, err)
		return
	}

	message := "Successfully created user"
	WriteJSONResponse(rw, http.StatusCreated, toUserResponse(*output), &message, nil)
}

func (a API) ListUsers(rw http.ResponseWriter, r *http.Request) {
	output, err := a.userService.List(r.Context(), services.ListUsersInput{})
	if err != nil {
		WriteJSONResponse(rw, statusForError(err), nil, nil, err)
		return
	}

	users := make([]UserResponse, 0, len(output.Users))
	for _, u := range output.Users {
		users = append(users, toUserResponse(u))
	}

	message := "Successfully listed users"
	WriteJSONResponse(rw, http.StatusOK, ListUsersResponse{Users: users}, &message, nil)
}

func (a API) GetUser(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}

	output, err := a.userService.Get(r.Context(), services.GetUserInput{Id: id})
	if err != nil {
		WriteJSONResponse(rw, statusForError(err), nil, nil, err)
		return
	}

	message := "Successfully retrieved user"
	WriteJSONResponse(rw, http.StatusOK, toUserResponse(*output), &message, nil)
}

func (a API) UpdateUser(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}

	r.Body = http.MaxBytesReader(rw, r.Body, 1<<20)
	req := UpdateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}
	if err := req.Validate(); err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}

	output, err := a.userService.Update(r.Context(), services.UpdateUserInput{
		Id:    id,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		WriteJSONResponse(rw, statusForError(err), nil, nil, err)
		return
	}

	message := "Successfully updated user"
	WriteJSONResponse(rw, http.StatusOK, toUserResponse(*output), &message, nil)
}

func (a API) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteJSONResponse(rw, http.StatusBadRequest, nil, nil, err)
		return
	}

	if err := a.userService.Delete(r.Context(), services.DeleteUserInput{Id: id}); err != nil {
		WriteJSONResponse(rw, statusForError(err), nil, nil, err)
		return
	}

	message := "Successfully deleted user"
	WriteJSONResponse(rw, http.StatusOK, map[string]string{"id": id.String()}, &message, nil)
}

func toUserResponse(u services.UserOutput) UserResponse {
	return UserResponse{
		Id:        u.Id,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
