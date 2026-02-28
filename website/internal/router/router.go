package router

import (
	"net/http"
	"strings"

	"github.com/N3moAhead/bombahead/website/internal/cfg"
	"github.com/N3moAhead/bombahead/website/internal/db"
	"github.com/N3moAhead/bombahead/website/internal/models"
	"github.com/N3moAhead/bombahead/website/internal/templates/home"
	"github.com/N3moAhead/bombahead/website/internal/templates/leaderboard"
	"github.com/N3moAhead/bombahead/website/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

var log = logger.New("[Router]")

var store sessions.Store

const appSessionName = "bomberman-session"

func Start(cfg *cfg.Config) {
	log.Info("Trusted origin: %s", cfg.BaseURL)

	authSetup(cfg)

	router := chi.NewRouter()

	// --- Middlewares ---
	router.Use(middleware.Logger)
	csrfMiddleware := csrf.Protect(
		[]byte(cfg.CSRFAuthKey),
		csrf.Secure(cfg.IsProduction),
	)
	router.Use(csrfMiddleware)
	router.Use(userMiddleware)

	// Serving static files
	FileServer(router, "/static", http.Dir("./static"))

	/// --- Auth Routes ---
	router.Get("/auth/{provider}", githubLogin)
	router.Get("/auth/{provider}/callback", githubLoginCallback)
	router.Post("/logout", logout)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(userContextKey).(*models.User)
		h := home.Home(csrf.Token(r), user)
		err := h.Render(r.Context(), w)
		renderError(err, w)
	})

	router.Get("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(userContextKey).(*models.User)
		bots, _ := db.GetLeaderboard(0, 50)
		s := leaderboard.Leaderboard(csrf.Token(r), user, bots)
		err := s.Render(r.Context(), w)
		renderError(err, w)
	})

	router.Mount("/matches", MatchRoutes())

	router.Route("/bots", botRoutes)
	router.Route("/docs", docRoutes)

	log.Info("Starting website on port %s", cfg.Port)
	err := http.ListenAndServe(cfg.Port, router)
	if err != nil {
		log.Error("failed to start website: %v", err)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not allow url parameters")
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
