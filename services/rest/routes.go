package rest

import "github.com/go-chi/chi"

func (r *REST) AddRoutes() {
	r.Router.Route("/", func(rt chi.Router) {
		rt.Get("/", r.GetRoot)
	})
}
