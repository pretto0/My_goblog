package controllers

import (
	"My_goblog/app/models/article"
	"My_goblog/pkg/logger"
    "My_goblog/pkg/route"
    "My_goblog/pkg/types"

    "text/template"
    "database/sql"
	"net/http"
    "fmt"
)

// ArticlesController 文章相关页面
type ArticlesController struct {
}

// Show 文章详情页面
func (*ArticlesController) Show(w http.ResponseWriter, r *http.Request) {
	id := route.GetRouteVariable("id", r)

    article, err := article.Get(id)

    if err != nil{
        if err == sql.ErrNoRows{
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprint(w,"404 文章未找到")
        }else{
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        }
    } else {
        tmpl, err := template.New("show.gohtml").
            Funcs(template.FuncMap{
                "RouteName2URL": route.RouteName2URL,
                "Uint64ToString": types.Uint64ToString,
            }).ParseFiles("resources/views/articles/show.gohtml")
        
        logger.LogError(err)

        err = tmpl.Execute(w, article)
        logger.LogError(err)
    }
}