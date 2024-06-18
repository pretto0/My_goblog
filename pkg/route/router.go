package route

import (
    "github.com/gorilla/mux"
    "net/http"
    
)

var route *mux.Router

func SetRoute(r *mux.Router) {
    route = r
}

func RouteName2URL(routeName string, pairs ...string) string {
    url, err := route.Get(routeName).URL(pairs...)
    if err != nil {
        // checkError(err)
        return ""
    }

    return url.String()              
}

func GetRouteVariable(parameterName string, r *http.Request) string {
    vars := mux.Vars(r)
    return vars[parameterName]
}