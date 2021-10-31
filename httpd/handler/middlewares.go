package handler

import (
	"context"
	"database/sql"
	"errors"
	"go-blog/platform/article"
	"go-blog/platform/comment"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"
	"strconv"
	"strings"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

const (
	TimeFormat    = "02-01-2006" // DD-MM-YYYY
	DateDelimiter = "|"
)

type key int

const (
	ArticleRepoKey key = 0
	ArticleKey     key = 1
	UserRepoKey    key = 2
	UserKey        key = 3
	RoleRepoKey    key = 4
	RoleKey        key = 5
	CommentRepoKey key = 6
	CommentKey     key = 7
	PageKey        key = 8
	DatesKey       key = 9
	UserIDKey      key = 10
	ClaimsKey      key = 11
)

func ProvideCommentRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := comment.NewRepo(db)
			ctx := context.WithValue(r.Context(), CommentRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideArticleRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := article.NewRepo(db)
			ctx := context.WithValue(r.Context(), ArticleRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideRoleRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := role.NewRepo(db)
			ctx := context.WithValue(r.Context(), RoleRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideUserRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := user.NewRepo(db)
			ctx := context.WithValue(r.Context(), UserRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(RoleRepoKey).(*role.Repo)

		var strRoleID string
		if strRoleID = chi.URLParam(r, "roleID"); strRoleID == "" {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing role ID")))
			return
		}

		roleID, err := strconv.ParseInt(strRoleID, 10, 64)
		if err != nil || roleID < 1 {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid role id.")))
			return
		}

		role, err := repo.GetByID(roleID)
		if err != nil {
			render.Render(w, r, status.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), RoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CommentIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(CommentRepoKey).(*comment.Repo)

		var strCommentID string
		if strCommentID = chi.URLParam(r, "commentID"); strCommentID == "" {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("missing comment ID")))
			return
		}

		commentID, err := strconv.ParseInt(strCommentID, 10, 64)
		if err != nil || commentID < 1 {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid comment id.")))
			return
		}

		comment, err := repo.GetByID(commentID)
		if err != nil {
			render.Render(w, r, status.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), CommentKey, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserSelfID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var strUserID string

		if strUserID = chi.URLParam(r, "userID"); strUserID == "" {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Missing user ID.")))
			return
		}

		userID, err := strconv.ParseInt(strUserID, 10, 64)
		if err != nil || userID < 1 {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid user id.")))
			return
		}

		roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
		claims := r.Context().Value(ClaimsKey).(user.Claims)
		userRole, err := roleRepo.GetByID(claims.RoleID)
		if err != nil {
			render.Render(w, r, status.ErrInternal(err))
			return
		}

		if userID != claims.UserID && !userRole.Check(role.CanManageOtherUsers) {
			render.Render(w, r, status.ErrUnauthorized("You are not the user."))
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ArticleIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(ArticleRepoKey).(*article.Repo)

		var articleID string
		if articleID = chi.URLParam(r, "articleID"); articleID == "" {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("missing article ID")))
			return
		}

		article, err := repo.GetByID(articleID)
		if err != nil {
			render.Render(w, r, status.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ArticleKey, article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pageNum int = 0

		if page := r.FormValue("page"); page != "" {
			var err error
			pageNum, err = strconv.Atoi(page)
			if err != nil || pageNum <= 0 {
				render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid page number.")))
				return
			}
		}

		ctx := context.WithValue(r.Context(), PageKey, pageNum)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ParseDate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var from, to int64
		var err error

		if date := r.FormValue("date"); date != "" {
			strDates := strings.Split(date, DateDelimiter)
			if len(strDates) != 2 {
				render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid date format.")))
				return
			}

			if from, err = strconv.ParseInt(strDates[0], 10, 64); err != nil {
				render.Render(w, r, status.ErrInvalidRequest(err))
				return
			}

			if to, err = strconv.ParseInt(strDates[1], 10, 64); err != nil {
				render.Render(w, r, status.ErrInvalidRequest(err))
				return
			}

			if from < 0 || to < from {
				render.Render(w, r, status.ErrInvalidRequest(errors.New("Invalid date format.")))
				return
			}
		}

		ctx := context.WithValue(r.Context(), DatesKey, [2]int64{from, to})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthenticatorNoPass(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil || jwt.Validate(token) != nil {
			render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKey, user.NewClaimsFromMap(claims))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthenticatorPass(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())
		var ctx context.Context

		if err != nil || token == nil || jwt.Validate(token) != nil {
			ctx = context.WithValue(r.Context(), ClaimsKey, user.NotAuthenticated)
		} else {
			ctx = context.WithValue(r.Context(), ClaimsKey, user.NewClaimsFromMap(claims))
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
