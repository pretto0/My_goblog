package main

import (
	"database/sql"
	// "errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"unicode/utf8"
    "My_goblog/pkg/route"
    "My_goblog/pkg/logger"
    "My_goblog/pkg/types"
    "My_goblog/pkg/database"
    "My_goblog/bootstrap"

    "github.com/gorilla/mux"
)

var router *mux.Router
var db *sql.DB

type ArticlesFormData struct{
    Title, Body string
    URL *url.URL
    Errors map[string]string
}

type Article struct{
    Title, Body string
    ID int64
}

func (a Article) Link() string {
    showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
    if err != nil {
        logger.LogError(err)
        return ""
    }
    return showURL.String()
}

func (a Article) Delete() (rowsAffected int64, err error) {
    rs, err := db.Exec("DELETE FROM articles WHERE id = " + strconv.FormatInt(a.ID, 10))
    if err != nil {
        return 0, err
    }
    if n, _ := rs.RowsAffected(); n > 0 {
        return n, nil
    }
    return 0,nil
}



func validateArticleFormData(title string, body string) map[string]string {
    errors := make(map[string]string)
    // 验证标题
    if title == "" {
        errors["title"] = "标题不能为空"
    } else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
        errors["title"] = "标题长度需介于 3-40"
    }

    // 验证内容
    if body == "" {
        errors["body"] = "内容不能为空"
    } else if utf8.RuneCountInString(body) < 10 {
        errors["body"] = "内容长度需大于或等于 10 个字节"
    }

    return errors
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {

    id := getRouteVariable("id", r)

    article, err := getArticleByID(id)

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
                "Int64ToString": types.Int64ToString,
            }).ParseFiles("resources/views/articles/show.gohtml")
        
        logger.LogError(err)

        err = tmpl.Execute(w, article)
        logger.LogError(err)
    }
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "访问文章列表")
    rows,err := db.Query("SELECT * from articles")
    logger.LogError(err)
    defer rows.Close()

    var articles []Article

    for rows.Next() {
        var article Article
        // 2.1 扫描每一行的结果并赋值到一个 article 对象中
        err := rows.Scan(&article.ID, &article.Title, &article.Body)
        logger.LogError(err)
        // 2.2 将 article 追加到 articles 的这个数组中
        articles = append(articles, article)
    }

    err = rows.Err()
    logger.LogError(err)

    tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
    logger.LogError(err)

    err = tmpl.Execute(w, articles)
    logger.LogError(err)

}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {

    title := r.PostFormValue("title")
    body := r.PostFormValue("body")

    errors := validateArticleFormData(title,body)

    if len(errors) == 0 {
        lastInsertID, err := saveArticleToDB(title, body)
        if lastInsertID > 0 {
            fmt.Fprint(w, "插入成功，ID 为"+strconv.FormatInt(lastInsertID, 10))
        } else {
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w,  "500 服务器内部错误")
        }
    } else {
        storeURL, _ := router.Get("articles.store").URL()
        data := ArticlesFormData{
            Title:  title,
            Body:   body,
            URL:    storeURL,
            Errors: errors,
        }
        tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
        if err != nil {
            panic(err)
        }

        err = tmpl.Execute(w, data)
        if err != nil {
            panic(err)
        }
    }
}

func saveArticleToDB(title string, body string)(int64, error){
    var (
        id int64
        err error
        rs sql.Result
        stmt *sql.Stmt
    )

    stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?,?)")
    if err != nil{
        return 0, err
    }

    defer stmt.Close()

    rs, err = stmt.Exec(title, body)
    if err != nil {
        return 0, err
    }

    if id, err = rs.LastInsertId(); id > 0 {
        return id, nil
    }

    return 0, err
}


func getRouteVariable(parameterName string, r *http.Request) string {
    vars := mux.Vars(r)
    return vars[parameterName]
}

func getArticleByID(id string) (Article, error) {
    article := Article{}
    query := "SELECT * FROM articles WHERE id = ?"
    err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
    return article, err
}


func articlesCreateHandler(w http.ResponseWriter, r *http.Request){
    storeURL, _ := router.Get("articles.store").URL()
    data := ArticlesFormData{
        Title:  "",
        Body:   "",
        URL:    storeURL,
        Errors: nil,
    }

    tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
    if err != nil {
        panic(err)
    }

    err = tmpl.Execute(w,data)

    if err !=nil {
        panic(err)
    }

}

func articlesEditHandler(w http.ResponseWriter, r *http.Request){

    id := getRouteVariable("id", r)

    article, err := getArticleByID(id)

    if err != nil {
        if err == sql.ErrNoRows {
            // 3.1 数据未找到
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprint(w, "404 文章未找到")
        } else {
            // 3.2 数据库错误
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        }
    } else {
        updateURL, _ := router.Get("articles.update").URL("id", id)

        data := ArticlesFormData{
            Title:  article.Title,
            Body:   article.Body,
            URL:    updateURL,
            Errors: nil,
        }
        tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
        logger.LogError(err)

        err = tmpl.Execute(w, data)
        logger.LogError(err)
    }
}

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request){
    id := getRouteVariable("id", r)
    _, err := getArticleByID(id)

    if err != nil {
        if err == sql.ErrNoRows {
            // 3.1 数据未找到
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprint(w, "404 文章未找到")
        } else {
            // 3.2 数据库错误
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        }
    }else{
        title := r.PostFormValue("title")
        body := r.PostFormValue("body")
        
        
        errors := validateArticleFormData(title,body)
        if len(errors) == 0 {

            query := "UPDATE articles SET title = ?, body = ? WHERE id = ?"

            rs, err := db.Exec(query,title,body,id)

            if err != nil {
                logger.LogError(err)
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprint(w, "500 服务器内部错误")
            }

            if n, _ := rs.RowsAffected();n>0{
                showURL, _ := router.Get("articles.show").URL("id", id)
                http.Redirect(w, r, showURL.String(), http.StatusFound)
            }else{
                fmt.Fprint(w, "您没有做任何更改！")
            }
            
            } else {

            updateURL,_ := router.Get("article.update").URL("id",id)
            data := ArticlesFormData{
                Title:  title,
                Body:   body,
                URL:    updateURL,
                Errors: errors,
            }
            tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
            logger.LogError(err)

            err = tmpl.Execute(w, data)
            logger.LogError(err)

        }
    }
}

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request){
    id := getRouteVariable("id", r)
    article, err := getArticleByID(id)
    if err != nil {
        if err == sql.ErrNoRows {
            // 3.1 数据未找到
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprint(w, "404 文章未找到")
        } else {
            // 3.2 数据库错误
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        }
    
    }else{
        rowsAffected, err := article.Delete()
        if err != nil {
            // 应该是 SQL 报错了
            logger.LogError(err)
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        } else {
            // 4.2 未发生错误
            if rowsAffected > 0 {
                // 重定向到文章列表页
                indexURL, _ := router.Get("articles.index").URL()
                http.Redirect(w, r, indexURL.String(), http.StatusFound)
            } else {
                // Edge case
                w.WriteHeader(http.StatusNotFound)
                fmt.Fprint(w, "404 文章未找到")
            }
    }
    }
}


func forceHTMLMiddleware(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        h.ServeHTTP(w, r)
    })
}

func removeTrailingSlash(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        if r.URL.Path != "/" {
            r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
        }

        next.ServeHTTP(w, r)
    })
}



func main() {
    database.Initialize()
    db = database.DB

    bootstrap.SetupDB()
    
    router = bootstrap.SetupRoute()

    router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
    router.HandleFunc("/articles/create", articlesStoreHandler).Methods("POST").Name("articles.store")
    router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
    router.HandleFunc("/articles/{id:[0-9]+}/edit",articlesEditHandler).Methods("GET").Name("article.edit")
    router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
    router.HandleFunc("/articles/index", articlesIndexHandler).Methods("GET").Name("articles.index")
    router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")


    router.Use(forceHTMLMiddleware)

    http.ListenAndServe(":3000", removeTrailingSlash(router))
}