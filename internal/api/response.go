package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/louispy/gotemplate/internal/custerr"
)

type jsonResponse struct {
	Data    any     `json:"data,omitempty"`
	Message *string `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func WriteJSONResponse(rw http.ResponseWriter, statusCode int, body any, message *string, err error) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	resp := jsonResponse{}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Data = body
		resp.Message = message
	}
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		log.Printf("Error encoding response as json: %v", err.Error())
	}
}

func statusForError(err error) int {
	switch err {
	case custerr.ErrDataNotFound:
		return http.StatusNotFound
	case custerr.ErrDuplicate:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

type UserResponse struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
}
