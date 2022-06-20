package rest

import "github.com/go-chi/chi"

func (rest *R) AddRoutes() {
	rest.Router.Route("/", func(router chi.Router) {
		router.Get("/", rest.GetRoot)

		router.Route("/api/v1", func(router chi.Router) {
			router.Use(rest.AuthMiddleware.CheckJWT)
			router.Get("/", rest.GetRoot)
		})
	})
}
