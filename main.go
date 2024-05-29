package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type ServiceAccountRequest struct {
	GlUrl       string `form:"glUrl" binding:"required"`
	GlToken     string `form:"glToken" binding:"required"`
	Name        string `form:"saName" binding:"required"`
	Displayname string `form:"saDisplayName" binding:"required"`
}

func RequestOptionFuncWithParam(params map[string]string) gitlab.RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		for k, v := range params {
			q := req.URL.Query()
			q.Add(k, v)
			req.URL.RawQuery = q.Encode()
		}
		return nil
	}
}

func Serviceaccount(c *gin.Context) {
	var sar ServiceAccountRequest
	if err := c.ShouldBind(&sar); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}
	git, err := gitlab.NewClient(sar.GlToken, gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", sar.GlUrl)))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}
	// check if user is admin
	user, _, err := git.Users.CreateServiceAccountUser(RequestOptionFuncWithParam(map[string]string{"name": sar.Displayname, "username": sar.Name}))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
		return
	}
	c.JSON(http.StatusOK, user)
	// c.Redirect(http.StatusFound, "/")
}

func main() {
	r := gin.Default()

	year, _, _ := time.Now().Date()
	component := Site("Index", year)
	hander := templ.Handler(component)

	r.GET("/", gin.WrapH(hander))
	r.POST("/serviceaccount", Serviceaccount)
	r.Run(":8080")
}
