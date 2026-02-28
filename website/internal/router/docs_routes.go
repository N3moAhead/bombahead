package router

import (
	"net/http"

	internaldocs "github.com/N3moAhead/bombahead/website/internal/docs"
	"github.com/N3moAhead/bombahead/website/internal/models"
	docstpl "github.com/N3moAhead/bombahead/website/internal/templates/docs"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func docRoutes(botRouter chi.Router) {
	botRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		renderDocBySlug(w, r, "")
	})

	botRouter.Get("/{slug}", func(w http.ResponseWriter, r *http.Request) {
		renderDocBySlug(w, r, chi.URLParam(r, "slug"))
	})
}

func renderDocBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	user, _ := r.Context().Value(userContextKey).(*models.User)

	normalizedSlug := slug
	if normalizedSlug == "" {
		normalizedSlug = "index"
	}

	htmlContent, err := internaldocs.LoadHTML(normalizedSlug)
	if err != nil {
		htmlContent, err = internaldocs.LoadHTML("index")
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	component := docstpl.Doc(user, csrf.Token(r), htmlContent, normalizedSlug)
	component.Render(r.Context(), w)
}
