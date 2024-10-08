package controllers

import (
	"My_goblog/app/models/article"
	"My_goblog/pkg/logger"
    "My_goblog/pkg/route"
    "My_goblog/pkg/types"

    "strconv"
    "unicode/utf8"
    "text/template"
    "database/sql"
	"net/http"
    "fmt"
    "gorm.io/gorm"
)

// ArticlesController 文章相关页面
type ArticlesController struct {
}

type ArticlesFormData struct {
    Title, Body string
    URL         string
    Errors      map[string]string
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

func (*ArticlesController) Index(w http.ResponseWriter, r *http.Request){
    articles, err := article.GetALL()

    if err != nil {
        logger.LogError(err)
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprint(w, "500 服务器内部错误")
    }else{
        tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
        logger.LogError(err)

        err = tmpl.Execute(w, articles)
        logger.LogError(err)
    }

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


func (*ArticlesController) Create(w http.ResponseWriter, r *http.Request){
    storeURL := route.RouteName2URL("articles.store")
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
    if err != nil {
        panic(err)
    }
}




func (*ArticlesController) Store(w http.ResponseWriter, r *http.Request){
    title := r.PostFormValue("title")
    body := r.PostFormValue("body")

    errors := validateArticleFormData(title, body)

    if len(errors) == 0 {
        _article := article.Article{
            Title: title,
            Body: body,
        }
        _article.Create()

        
        if _article.ID > 0 {
            fmt.Fprint(w, "插入成功，ID 为"+strconv.FormatUint(_article.ID, 10))
        } else {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "创建文章失败，请联系管理员")
        }
    } else {
        storeURL := route.RouteName2URL("articles.store")
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

func (*ArticlesController) Edit(w http.ResponseWriter, r *http.Request){
    id := route.GetRouteVariable("id", r)

    article, err := article.Get(id)

    if err != nil {
        if err == gorm.ErrRecordNotFound {
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
        updateURL := route.RouteName2URL("articles.update","id", id)

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


func (*ArticlesController) Update(w http.ResponseWriter, r *http.Request){
    id := route.GetRouteVariable("id", r)

    _article, err := article.Get(id)

    if err != nil {
        if err == gorm.ErrRecordNotFound {
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
        title := r.PostFormValue("title")
        body := r.PostFormValue("body")
        
        errors := validateArticleFormData(title, body)

        if len(errors) == 0 {
            _article.Title = title
            _article.Body = body

            rowsAffected, err := _article.Update()

            if err != nil {
                // 数据库错误
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprint(w, "500 服务器内部错误")
                return
            }
            if rowsAffected > 0 {
                showURL := route.RouteName2URL("articles.show", "id", id)
                http.Redirect(w, r, showURL, http.StatusFound)
            } else {
                fmt.Fprint(w, "您没有做任何更改！")
            }
        }else{
            updateURL := route.RouteName2URL("articles.update", "id", id)
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


func (*ArticlesController) Delete(w http.ResponseWriter, r *http.Request) {

    // 1. 获取 URL 参数
    id := route.GetRouteVariable("id", r)

    // 2. 读取对应的文章数据
    _article, err := article.Get(id)

    // 3. 如果出现错误
    if err != nil {
        if err == gorm.ErrRecordNotFound {
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
        // 4. 未出现错误，执行删除操作
        rowsAffected, err := _article.Delete()

        // 4.1 发生错误
        if err != nil {
            // 应该是 SQL 报错了
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "500 服务器内部错误")
        } else {
            // 4.2 未发生错误
            if rowsAffected > 0 {
                // 重定向到文章列表页
                indexURL := route.RouteName2URL("articles.index")
                http.Redirect(w, r, indexURL, http.StatusFound)
            } else {
                // Edge case
                w.WriteHeader(http.StatusNotFound)
                fmt.Fprint(w, "404 文章未找到")
            }
        }
    }
}