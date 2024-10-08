package routes

import (
	"My_goblog/app/http/controllers"
	"My_goblog/app/http/middlewares"
	// "My_goblog/app/http/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterWebRoutes 注册网页相关路由
func RegisterWebRoutes(r *mux.Router) {

	pc := new(controllers.PagesController)

    // 静态页面
    r.HandleFunc("/", pc.Home).Methods("GET").Name("home")
    r.HandleFunc("/about", pc.About).Methods("GET").Name("about")
    r.NotFoundHandler = http.HandlerFunc(pc.NotFound)

	ac := new(controllers.ArticlesController)

	r.HandleFunc("/articles/{id:[0-9]+}", ac.Show).Methods("GET").Name("articles.show")
	r.HandleFunc("/articles", ac.Index).Methods("GET").Name("articles.index")
	r.HandleFunc("/articles/create", ac.Create).Methods("GET").Name("articles.create")
	r.HandleFunc("/articles/create", ac.Store).Methods("POST").Name("articles.store")
	r.HandleFunc("/articles/{id:[0-9]+}/edit",ac.Edit).Methods("GET").Name("article.edit")
	r.HandleFunc("/articles/{id:[0-9]+}", ac.Update).Methods("POST").Name("articles.update")
	r.HandleFunc("/articles/{id:[0-9]+}/delete", ac.Delete).Methods("POST").Name("articles.delete")
	
	//静态资源
	// r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./public/css"))))
	r.PathPrefix("/css/").Handler(middlewares.ForceCSS(http.FileServer(http.Dir("./public"))))
    r.PathPrefix("/js/").Handler(http.FileServer(http.Dir("./public")))

	

	// r.Use(middlewares.ForceHTML)
}