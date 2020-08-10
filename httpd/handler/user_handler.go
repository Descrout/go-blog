package handler

import (
	"errors"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

func UserDelete(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := repo.Delete(userTemp.ID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	userPayload := user.NewUserPayload(userTemp, roleRepo)

	if err := render.Bind(r, userPayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	userTemp = userPayload.User
	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := repo.Update(userTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
}

func UserGetByID(w http.ResponseWriter, r *http.Request) {
	userTemp := r.Context().Value(UserKey).(*user.User)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
}

func UserGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	users := repo.GetAll()
	_ = render.RenderList(w, r, user.NewUserListPayload(users, roleRepo))
}

func UserLoginPost(tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &user.UserPayload{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, status.ErrInvalidRequest(err))
			return
		}
		userTemp := data.User

		if isValid := user.EmailRegex.MatchString(userTemp.Email); !isValid {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid e-mail.")))
			return
		}
		if isValid := user.PasswordRegex.MatchString(userTemp.Password); !isValid {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid password.")))
			return
		}

		repo := r.Context().Value(UserRepoKey).(*user.Repo)

		resultUser, err := repo.GetByEmail(userTemp.Email)
		if err != nil {
			render.Render(w, r, status.ErrUnauthorized("Wrong Credentials."))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(resultUser.Password), []byte(userTemp.Password))
		if err == nil {
			roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
			userData := user.NewUserPayload(resultUser, roleRepo)
			claims := jwt.MapClaims{"user_id": userData.ID, "role_id": userData.Role_ID}

			var expiration time.Time

			if r.FormValue("remember") == "1" {
				expiration = time.Now().Add(365 * 24 * time.Hour)
			} else {
				expiration = time.Now().Add(time.Hour)
			}

			jwtauth.SetExpiry(claims, expiration)
			_, tokenString, _ := tokenAuth.Encode(claims)
			userData.Token = tokenString

			http.SetCookie(w, &http.Cookie{Name: "jwt", Value: tokenString, Expires: expiration, HttpOnly: true})
			render.Status(r, http.StatusOK)
			render.Render(w, r, userData)
		} else {
			render.Render(w, r, status.ErrUnauthorized("Wrong Credentials."))
			return
		}
	}
}

func UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	data := &user.UserPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}
	userTemp := data.User

	if isValid := user.NameRegex.MatchString(userTemp.Name); !isValid {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid name.")))
		return
	}
	if isValid := user.EmailRegex.MatchString(userTemp.Email); !isValid {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid e-mail.")))
		return
	}
	if isValid := user.PasswordRegex.MatchString(userTemp.Password); !isValid {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid password.")))
		return
	}

	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if exist, err := repo.DoesEmailExist(userTemp.Email); err == nil {
		if exist {
			render.Render(w, r, status.ErrConflict("Email already registered."))
			return
		}
	} else {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userTemp.Password), bcrypt.DefaultCost); err == nil {
		userTemp.Password = string(hashedPassword)
	} else {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if id, err := repo.Add(userTemp); err == nil {
		data.User.ID = id
		data.User.Role_ID = 1
	} else {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	render.Status(r, http.StatusCreated)
	render.Render(w, r, user.NewUserPayload(data.User, roleRepo))
}