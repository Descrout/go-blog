package handler

import (
	"errors"
	errs "go-blog/platform/errors"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

func UserDelete(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := repo.Delete(userTemp.ID); err != nil {
		render.Render(w, r, errs.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(userTemp))
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	userPayload := user.NewUserPayload(userTemp)

	if err := render.Bind(r, userPayload); err != nil {
		render.Render(w, r, errs.ErrInvalidRequest(err))
		return
	}

	userTemp = userPayload.User
	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := repo.Update(userTemp); err != nil {
		render.Render(w, r, errs.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(userTemp))
}

func UserGetByID(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	render.Render(w, r, user.NewUserPayload(userTemp))
}

func UserGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(UserRepoKey).(*user.Repo)
	users := repo.GetAll()
	render.RenderList(w, r, user.NewUserListPayload(users))
}

func UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	data := &user.UserPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errs.ErrInvalidRequest(err))
		return
	}
	userTemp := data.User

	if isValid := user.NameRegex.MatchString(userTemp.Name); !isValid {
		render.Render(w, r, errs.ErrInvalidRequest(errors.New("Invalid name.")))
		return
	}
	if isValid := user.EmailRegex.MatchString(userTemp.Email); !isValid {
		render.Render(w, r, errs.ErrInvalidRequest(errors.New("Invalid e-mail.")))
		return
	}
	if isValid := user.PasswordRegex.MatchString(userTemp.Password); !isValid {
		render.Render(w, r, errs.ErrInvalidRequest(errors.New("Invalid password.")))
		return
	}

	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if exist, err := repo.DoesEmailExist(userTemp.Email); err == nil {
		if exist {
			render.Render(w, r, errs.ErrInvalidRequest(errors.New("Email already registered.")))
			return
		}
	} else {
		render.Render(w, r, errs.ErrInternal(err))
		return
	}

	if hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userTemp.Password), bcrypt.DefaultCost); err == nil {
		userTemp.Password = string(hashedPassword)
	} else {
		render.Render(w, r, errs.ErrInternal(err))
		return
	}

	if id, err := repo.Add(userTemp); err == nil {
		data.User.ID = id
	} else {
		render.Render(w, r, errs.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, data)
}
