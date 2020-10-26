package user

import (
	"errors"
	"go-blog/platform/role"
	"net/http"
	"regexp"

	"github.com/go-chi/render"
)

var NameRegex = regexp.MustCompile(`^[a-zA-Z]+((([',. -][a-zA-Z ])?[a-zA-Z]){2,20})$`)
var EmailRegex = regexp.MustCompile(`(.+)@(.+){2,}\.(.+){2,}`)
var PasswordRegex = regexp.MustCompile(`^(\S{6,20})$`)

type UpdateEmail struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (e *UpdateEmail) Bind(r *http.Request) error {
	if e.Password == "" {
		return errors.New("Missing password field")
	}
	if e.Email == "" {
		return errors.New("Missing email field")
	}
	if !EmailRegex.MatchString(e.Email) {
		return errors.New("Email requirements does not match.")
	}
	return nil
}

type UpdatePassword struct {
	OldPassword string `json:"oldPassword"`
	Password    string `json:"password"`
}

func (p *UpdatePassword) Bind(r *http.Request) error {
	if p.Password == "" {
		return errors.New("Missing password field")
	}
	if p.OldPassword == "" {
		return errors.New("Missing old password field")
	}
	if !PasswordRegex.MatchString(p.Password) {
		return errors.New("Password requirements does not match.")
	}

	if p.Password == p.OldPassword {
		return errors.New("Old password must be different than current password.")
	}

	return nil
}

type User struct {
	ID       int64  `json:"id"`
	Role_ID  int64  `json:"-"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
	Image    string `json:"image,omitempty"`
}

type UserPayload struct {
	*User
	Role  *role.RolePayload `json:"role"`
	Token string            `json:"token,omitempty"`
}

func NewUserPayload(user *User, roleRepo *role.Repo) *UserPayload {
	uPayload := &UserPayload{User: user}
	if uPayload.Role == nil && roleRepo != nil {
		if roleTemp, err := roleRepo.GetByID(user.Role_ID); err == nil {
			uPayload.Role = role.NewRolePayload(roleTemp)
		}
	}
	return uPayload
}

func NewUserListPayload(users []*User, roleRepo *role.Repo) []render.Renderer {
	list := []render.Renderer{}
	for _, user := range users {
		list = append(list, NewUserPayload(user, roleRepo))
	}
	return list
}

func (u *UserPayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if u.User == nil {
		return errors.New("missing required User fields.")
	}
	u.Token = ""
	return nil
}

func (u *UserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	u.Password = ""
	return nil
}
