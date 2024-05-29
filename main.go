package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type ServiceAccountRequest struct {
	GlUrl       string `form:"glUrl" binding:"required"`
	GlToken     string `form:"glToken" binding:"required"`
	Name        string `form:"saName"`
	Displayname string `form:"saDisplayName"`
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

func Out(c *gin.Context, in string, b *bytes.Buffer) {
	Output(in).Render(context.Background(), b)
	c.String(200, b.String())
}

func Serviceaccount(c *gin.Context) {
	var sar ServiceAccountRequest
	params := map[string]string{}
	buf := new(bytes.Buffer)
	if err := c.ShouldBind(&sar); err != nil {
		Out(c, err.Error(), buf)
		return
	}
	git, err := gitlab.NewClient(sar.GlToken, gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", sar.GlUrl)))
	if err != nil {
		Out(c, err.Error(), buf)
		return
	}
	currentUser, _, err := git.Users.CurrentUser(RequestOptionFuncWithParam(params))
	if err != nil {
		Out(c, err.Error(), buf)
		return
	}
	if !currentUser.IsAdmin {
		Out(c, "err: current user doesn't have admin permissions", buf)
		return
	}
	if len(sar.Displayname) > 0 {
		params = lo.Assign(params, map[string]string{"name": sar.Displayname})
	}
	if len(sar.Name) > 0 {
		params = lo.Assign(params, map[string]string{"username": sar.Name})
	}
	user, _, err := git.Users.CreateServiceAccountUser(RequestOptionFuncWithParam(params))
	if err != nil {
		Out(c, err.Error(), buf)
		return
	}
	Out(c, user.Username, buf)
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
