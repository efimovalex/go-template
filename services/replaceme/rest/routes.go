package rest

import "github.com/go-chi/chi"

func (rest *REST) AddRoutes() {
	rest.Router.Route("/", func(router chi.Router) {
		router.Get("/", rest.GetRoot)
	})
}
