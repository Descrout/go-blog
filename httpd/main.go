package main

import (
	"database/sql"
	"go-blog/httpd/handler"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	_ "github.com/mattn/go-sqlite3"
)

const (
	port        = ":3000"
	dbName      = "./blog.db"
	tokenSecret = "my_secret"
	servePath   = "static"
)

func main() {
	//Setup db
	db := setupDB(dbName)
	defer db.Close()

	//Create jwt authorization token
	tokenAuth := jwtauth.New("HS256", []byte(tokenSecret), nil)

	//Create a router
	r := chi.NewRouter()

	//Good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TODO-Index-Page"))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Your page is in another castle."))
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.ProvideUserRepo(db))
		r.Use(handler.ProvideRoleRepo(db))
		r.Route("/login", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("TODO-Login-Page"))
			})
			r.Post("/", handler.UserLoginPost(tokenAuth))
		})

		r.Route("/register", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("TODO-Register-Page"))
			})
			r.Post("/", handler.UserRegisterPost)
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(handler.ProvideArticleRepo(db))
		r.Use(handler.ProvideUserRepo(db))
		r.Use(handler.ProvideRoleRepo(db))

		r.Use(jwtauth.Verifier(tokenAuth)) // inits auth but does not check yet

		r.Route("/users", func(r chi.Router) {
			r.Get("/", handler.UserGetAll)

			r.Route("/{userID}", func(r chi.Router) {
				r.Get("/", handler.UserGetByID)

				r.With(jwtauth.Authenticator).Put("/role", handler.AssignRole)

				r.Group(func(r chi.Router) {
					r.Use(jwtauth.Authenticator, handler.UserAuthContext)
					r.Put("/name", handler.UserUpdateName)
					r.Put("/password", handler.UserUpdatePassword)
					//r.Put("/email", handler.UserUpdateEmail)
					r.Post("/image", handler.UserUpdateImage)
					r.Delete("/", handler.UserDelete)
				})
			})
		})

		r.Route("/roles", func(r chi.Router) {
			r.Get("/", handler.RoleGetAll)

			r.With(jwtauth.Authenticator).Post("/", handler.RolePost)

			r.Route("/{roleID}", func(r chi.Router) {
				r.Use(handler.RoleIDContext)
				r.Get("/", handler.RoleGetByID)
				r.With(jwtauth.Authenticator).Put("/", handler.RoleUpdate)
				r.With(jwtauth.Authenticator).Delete("/", handler.RoleDelete)
			})
		})

		r.Route("/comments", func(r chi.Router) {
			r.Use(handler.ProvideCommentRepo(db))

			r.Route("/id/{commentID}", func(r chi.Router) {
				r.Use(jwtauth.Authenticator)
				r.Use(handler.CommentIDContext)

				r.Put("/", handler.CommentUpdate)
				r.Delete("/", handler.CommentDelete)
			})

			r.Route("/{articleID}", func(r chi.Router) {
				r.Use(handler.ArticleIDContext)
				r.Get("/", handler.CommentsGet)
				r.With(jwtauth.Authenticator).Post("/", handler.CommentPost)
			})
		})

		r.Route("/articles", func(r chi.Router) {
			r.With(handler.Paginate).Get("/", handler.ArticleGetAll)
			//r.Get("/{month}-{day}-{year}", listArticlesByDate)
			//r.Get("/search", searchArticles)
			r.With(jwtauth.Authenticator).Post("/", handler.ArticlePost)

			r.Route("/{articleID}", func(r chi.Router) {
				r.Use(handler.ArticleIDContext)
				r.Get("/", handler.ArticleGetByID)
				r.With(jwtauth.Authenticator).Put("/", handler.ArticleUpdate)
				r.With(jwtauth.Authenticator).Delete("/", handler.ArticleDelete)
			})
		})
	})

	FileServer(r, handler.SERVE_PATH)

	log.Println("Serving on port " + port)
	http.ListenAndServe(port, r)
}

func FileServer(r chi.Router, path string) {
	workDir, _ := os.Getwd()
	root := http.Dir(filepath.Join(workDir, path))

	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func setupDB(filename string) *sql.DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "users" (
		"id"	INTEGER NOT NULL UNIQUE,
		"role_id"	INTEGER NOT NULL DEFAULT 1,
		"name"	TEXT NOT NULL,
		"password"	TEXT NOT NULL,
		"email" TEXT NOT NULL,
		"image"	TEXT NOT NULL DEFAULT "user.png",
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	CREATE TABLE IF NOT EXISTS "roles" (
		"id"	INTEGER NOT NULL UNIQUE,
		"name"	TEXT NOT NULL,
		"code"	INTEGER NOT NULL DEFAULT 1,
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	CREATE TABLE IF NOT EXISTS "articles" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"title"	TEXT NOT NULL,
		"body"	TEXT NOT NULL,
		"created_at"	INTEGER NOT NULL,
		"updated_at"	INTEGER NOT NULL,
		PRIMARY KEY("ID" AUTOINCREMENT)
	);
	CREATE TABLE IF NOT EXISTS "comments" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER,
		"article_id"	INTEGER,
		"body"	TEXT,
		"created_at"	INTEGER NOT NULL,
		"updated_at"	INTEGER NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	REPLACE INTO roles (id, name) values (1, "Guest");
	REPLACE INTO roles (id, name) values (2, "Author");
	REPLACE INTO roles (id, name, code) values (3, "Admin", 127);`)

	if err != nil {
		log.Fatal(err)
	}

	return db
}
