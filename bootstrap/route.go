package bootstrap

import (
	"My_goblog/routes"
	"My_goblog/pkg/route"

	"github.com/gorilla/mux"
)

func SetupRoute() *mux.Router {
	router := mux.NewRouter()
	routes.RegisterWebRoutes(router)

	route.SetRoute(router)

	return router
}