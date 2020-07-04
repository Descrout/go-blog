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
