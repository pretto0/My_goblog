// Package article 应用的文章模型
package article

import (
    "My_goblog/pkg/route"
    "strconv"
)

// Article 文章模型
type Article struct {
    ID    uint64 
    Title string
    Body  string
}

func (a Article) Link() string {
    
    return route.RouteName2URL("articles.show","id", strconv.FormatUint(a.ID, 10))
}
