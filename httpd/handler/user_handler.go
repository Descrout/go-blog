package handler

import (
	"errors"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

const (
	PROFILE_PICS = "/profile-pics"
	SERVE_PATH   = "/static"
	DEFAULT_PIC  = SERVE_PATH + PROFILE_PICS + "/user.png"
)

func UserDelete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserKey).(string)
	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := repo.Delete(userID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func AssignRole(w http.ResponseWriter, r *http.Request) {
	var userID string

	if userID = chi.URLParam(r, "userID"); userID == "" {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing user ID")))
		return
	}

	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	var roleID string
	if roleID = r.FormValue("id"); roleID == "" {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing role ID")))
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	if userRole, err := roleRepo.GetByID(claims["role_id"]); err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	} else if !userRole.Check(role.CanManageRole) {
		render.Render(w, r, status.ErrUnauthorized("You are not authorized to assign a role."))
		return
	}

	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)

	if err := userRepo.Update(userID, "role_id", roleID); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if userTemp, err := userRepo.GetByID(userID); err != nil {
		render.Render(w, r, status.ErrNotFound)
	} else {
		render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
	}
}

// only accepts 2mb png and jpeg
func UserUpdateImage(w http.ResponseWriter, r *http.Request) {
	max := int64(2 << 20)

	r.Body = http.MaxBytesReader(w, r.Body, max)

	err := r.ParseMultipartForm(max)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	file, _, err := r.FormFile("img")
	if err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}
	defer file.Close()

	workDir, _ := os.Getwd()
	picsDir := filepath.Join(workDir, SERVE_PATH, PROFILE_PICS)
	tempFile, err := ioutil.TempFile(picsDir, "profile-*.png")
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		os.Remove(tempFile.Name())
		return
	}

	fileType := http.DetectContentType(fileBytes)

	if fileType != "image/jpeg" && fileType != "image/png" {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid file type.")))
		os.Remove(tempFile.Name())
		return
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		os.Remove(tempFile.Name())
		return
	}

	userID := r.Context().Value(UserKey).(string)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	userTemp, err := userRepo.GetByID(userID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		os.Remove(tempFile.Name())
		return
	}

	imagePath := filepath.Join(SERVE_PATH, PROFILE_PICS, filepath.Base(tempFile.Name()))

	err = userRepo.Update(userID, "image", imagePath)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		os.Remove(tempFile.Name())
		return
	}

	if userTemp.Image != DEFAULT_PIC {
		os.Remove(filepath.Join(workDir, userTemp.Image))
	}

	userTemp.Image = imagePath
	render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
}

func UserUpdateEmail(w http.ResponseWriter, r *http.Request) {
	var data user.UpdateEmail

	if err := render.Bind(r, &data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)

	exist, err := userRepo.DoesEmailExist(data.Email)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}
	if exist {
		render.Render(w, r, status.ErrConflict("Email already registered."))
		return
	}

	userID := r.Context().Value(UserKey).(string)
	tempUser, _ := userRepo.GetByID(userID)

	err = bcrypt.CompareHashAndPassword([]byte(tempUser.Password), []byte(data.Password))
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Password is wrong."))
		return
	}

	if err = userRepo.Update(userID, "email", data.Email); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	tempUser.Email = data.Email

	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(tempUser, roleRepo))
}

func UserUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var data user.UpdatePassword

	if err := render.Bind(r, &data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	userID := r.Context().Value(UserKey).(string)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	tempUser, _ := userRepo.GetByID(userID)

	err := bcrypt.CompareHashAndPassword([]byte(tempUser.Password), []byte(data.OldPassword))
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Current password is wrong."))
		return
	}

	if hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost); err == nil {
		data.Password = string(hashedPassword)
	} else {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if err = userRepo.Update(userID, "password", data.Password); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(tempUser, roleRepo))
}

func UserUpdateName(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserKey).(string)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	var name string
	if name = r.FormValue("name"); name == "" {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing name field")))
		return
	}

	if !user.NameRegex.MatchString(name) {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid name.")))
		return
	}

	if err := userRepo.Update(userID, "name", name); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	userTemp, err := userRepo.GetByID(userID)
	if err != nil {
		render.Render(w, r, status.ErrNotFound)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
}

func UserGetByID(w http.ResponseWriter, r *http.Request) {
	var userID string

	if userID = chi.URLParam(r, "userID"); userID == "" {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing user ID")))
		return
	}

	repo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	userTemp, err := repo.GetByID(userID)
	if err != nil {
		render.Render(w, r, status.ErrNotFound)
		return
	}

	render.Render(w, r, user.NewUserPayload(userTemp, roleRepo))
}

func UserGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	users := repo.GetAll()
	render.RenderList(w, r, user.NewUserListPayload(users, roleRepo))
}

func UserLoginPost(tokenAuth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &user.UserPayload{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, status.ErrInvalidRequest(err))
			return
		}
		userTemp := data.User

		if !user.EmailRegex.MatchString(userTemp.Email) {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid e-mail.")))
			return
		}
		if !user.PasswordRegex.MatchString(userTemp.Password) {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Password requirements does not match.")))
			return
		}

		repo := r.Context().Value(UserRepoKey).(*user.Repo)

		resultUser, err := repo.GetByEmail(userTemp.Email)
		if err != nil {
			render.Render(w, r, status.ErrUnauthorized("Wrong Credentials."))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(resultUser.Password), []byte(userTemp.Password))
		if err != nil {
			render.Render(w, r, status.ErrUnauthorized("Wrong Credentials."))
			return
		}

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
	}
}

func UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	data := &user.UserPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}
	userTemp := data.User

	if !user.NameRegex.MatchString(userTemp.Name) {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid name.")))
		return
	}
	if !user.EmailRegex.MatchString(userTemp.Email) {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid e-mail.")))
		return
	}
	if !user.PasswordRegex.MatchString(userTemp.Password) {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("Password requirements does not match.")))
		return
	}

	repo := r.Context().Value(UserRepoKey).(*user.Repo)

	exist, err := repo.DoesEmailExist(userTemp.Email)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}
	if exist {
		render.Render(w, r, status.ErrConflict("Email already registered."))
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
