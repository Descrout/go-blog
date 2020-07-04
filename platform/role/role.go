package role

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code int64  `json:"code"`
}

type RolePayload struct {
	*Role
}

func NewRolePayload(role *Role) *RolePayload {
	return &RolePayload{Role: role}
}

func NewRoleListPayload(roles []*Role) []render.Renderer {
	list := []render.Renderer{}
	for _, role := range roles {
		list = append(list, NewRolePayload(role))
	}
	return list
}

func (rp *RolePayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if rp.Role == nil {
		return errors.New("missing required Role fields.")
	}
	return nil
}

func (rp *RolePayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	return nil
}
