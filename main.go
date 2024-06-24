package main

import (
	"net/http"
    "My_goblog/pkg/logger"

    "My_goblog/bootstrap"
    "My_goblog/app/http/middlewares"

)


func main() {

    bootstrap.SetupDB()
    
    router := bootstrap.SetupRoute()

    err := http.ListenAndServe(":3000", middlewares.RemoveTrailingSlash(router))
    logger.LogError(err)
}