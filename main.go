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

type TokenRequest struct {
	GlUrl     string   `form:"glUrl" binding:"required"`
	GlToken   string   `form:"glToken" binding:"required"`
	UserID    int      `form:"UserID" binding:"required"`
	Name      string   `form:"Name" binding:"required"`
	ExpiresAt string   `form:"ExpiresAt" binding:"required"`
	Scopes    []string `form:"Scopes" binding:"required"`
}

var Scopes = []string{
	"api",
	"read_user",
	"read_api",
	"read_repository",
	"write_repository",
	"read_registry",
	"write_registry",
	"sudo",
	"admin_mode",
	"create_runner",
	"ai_features",
	"k8s_proxy",
	"read_service_ping",
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

func OutToken(c *gin.Context, in string, b *bytes.Buffer) {
	OutputToken(in).Render(context.Background(), b)
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
	Out(c, fmt.Sprintf("ServiceUser %s with ID: %d created.", user.Username, user.ID), buf)
}

func Token(c *gin.Context) {
	var tr TokenRequest
	params := map[string]string{}
	buf := new(bytes.Buffer)
	if err := c.ShouldBind(&tr); err != nil {
		Out(c, err.Error(), buf)
		return
	}
	git, err := gitlab.NewClient(tr.GlToken, gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", tr.GlUrl)))
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
	gitlabIsoTime, err := gitlab.ParseISOTime(tr.ExpiresAt)
	if err != nil {
		Out(c, err.Error(), buf)
		return
	}
	opts := &gitlab.CreatePersonalAccessTokenOptions{
		Name:      &tr.Name,
		ExpiresAt: &gitlabIsoTime,
		Scopes:    &tr.Scopes,
	}
	pat, _, err := git.Users.CreatePersonalAccessToken(tr.UserID, opts)
	if err != nil {
		Out(c, err.Error(), buf)
		return
	}
	OutToken(c, pat.Token, buf)
}

func main() {
	r := gin.Default()

	year, _, _ := time.Now().Date()
	site := Site(year)
	sa := CreateServiceUser()
	pat := CreatePersonalAccessToken(Scopes)
	home := Home()

	r.GET("/", gin.WrapH(templ.Handler(site)))
	r.GET("/home", gin.WrapH(templ.Handler(home)))
	r.GET("/serviceaccount", gin.WrapH(templ.Handler(sa)))
	r.GET("/token", gin.WrapH(templ.Handler(pat)))

	r.POST("/serviceaccount", Serviceaccount)
	r.POST("/token", Token)

	r.Run(":8080")
}
