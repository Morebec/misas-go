// IMPORTANT: This file was auto-generated by the morebec/spectool program. Do not edit manually.

package user

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/morebec/misas-go/misas/command"
	"github.com/morebec/misas-go/misas/domain"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/httpapi"
	"net/http"
)

// httpUserCreate Allows creating a user
func httpUserCreate(r chi.Router, bus command.Bus) {
	r.Get("/user/create", func(w http.ResponseWriter, r *http.Request) {
		handleError := func(w http.ResponseWriter, r *http.Request, err error) {
			if !domain.IsDomainError(err) {
				w.WriteHeader(500)
				render.JSON(w, r, NewInternalError(err))
			}

			derr := err.(domain.Error)

			conv := map[domain.ErrorTypeName]int{
				// Returned when a user was not found.
				"user_not_found": 404,
			}
			c := conv[derr.TypeName()]
			w.WriteHeader(c)
			render.JSON(w, r, NewErrorResponse(derr.TypeName(), derr.Error(), derr.Data()))
		}

		// Decode request payload
		var input CreateUserCommand
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			w.WriteHeader(500)
			render.JSON(w, r, httpapi.NewInternalError(err))
			return
		}

		// Send to Domain Layer
		output, err := bus.Send(r.Context(), input)
		if err != nil {
			w.WriteHeader(400)
			render.JSON(w, r, httpapi.NewErrorResponse(output))
			return
		}

		render.JSON(w, r, httpapi.NewSuccessResponse(output))
	})
}

const CreateUserCommandTypeName command.TypeName = "user_account_management.user.create"

// CreateUserCommand Allows creating a user.
// In the system.
type CreateUserCommand struct {

	// Email address of the user
	// NOTE: This field contains personal data
	EmailAddress string `json:"emailAddress"`

	// ID of the user
	ID string `json:"id"`

	// list of permissions
	Permissions map[float64]string `json:"permissions"`

	//
	RefereeID *string `json:"refereeId"`

	// Registration of user
	Registration RenameUserCommand `json:"registration"`
}

func (c CreateUserCommand) TypeName() command.TypeName {
	return CreateUserCommandTypeName
}

// Type Represents the types of users.
type Type string

const (
	// Admin Represents a admin user
	Admin string = "ADMIN"

	// Normal Represents a normal user
	Normal string = "NORMAL"
)
