package handler

import (
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/render"
)

func RoleDelete(w http.ResponseWriter, r *http.Request) {
	roleTemp := r.Context().Value(RoleKey).(*role.Role)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	if userRole, err := roleRepo.GetByID(claims.RoleID); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else if !userRole.Check(role.CanManageRole) {
		render.Render(w, r, status.ErrUnauthorized("You are not authorized to manage roles."))
		return
	}

	if err := roleRepo.Delete(roleTemp.ID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func RoleUpdate(w http.ResponseWriter, r *http.Request) {
	roleTemp := r.Context().Value(RoleKey).(*role.Role)
	rolePayload := role.NewRolePayload(roleTemp)

	if err := render.Bind(r, rolePayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}
	roleTemp = rolePayload.Role
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	if userRole, err := roleRepo.GetByID(claims.RoleID); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else if !userRole.Check(role.CanManageRole) {
		render.Render(w, r, status.ErrUnauthorized("You are not authorized to manage roles."))
		return
	}

	if err := roleRepo.Update(roleTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, role.NewRolePayload(roleTemp))
}

func RoleGetByID(w http.ResponseWriter, r *http.Request) {
	roleTemp := r.Context().Value(RoleKey).(*role.Role)
	render.Render(w, r, role.NewRolePayload(roleTemp))
}

func RoleGetAll(w http.ResponseWriter, r *http.Request) {
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	roles := roleRepo.GetAll()
	render.RenderList(w, r, role.NewRoleListPayload(roles))
}

func RolePost(w http.ResponseWriter, r *http.Request) {
	data := &role.RolePayload{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	roleTemp := data.Role
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	if userRole, err := roleRepo.GetByID(claims.RoleID); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else if !userRole.Check(role.CanManageRole) {
		render.Render(w, r, status.ErrUnauthorized("You are not authorized to manage roles."))
		return
	}

	if id, err := roleRepo.Add(roleTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else {
		roleTemp.ID = id
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, role.NewRolePayload(roleTemp))
}
