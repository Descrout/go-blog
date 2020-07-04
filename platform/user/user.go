package user

import (
	"errors"
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
}

func NewUserPayload(user *User) *UserPayload {
	return &UserPayload{User: user}
}

func NewUserListPayload(users []*User) []render.Renderer {
	list := []render.Renderer{}
	for _, user := range users {
		list = append(list, NewUserPayload(user))
	}
	return list
}

func (u *UserPayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if u.User == nil {
		return errors.New("missing required User fields.")
	}
	return nil
}

func (u *UserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	u.Password = ""
	return nil
}
