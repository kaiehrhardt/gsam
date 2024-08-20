package main

import (
	"bytes"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

type TokenRequest struct {
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
	"manage_runner",
	"ai_features",
	"k8s_proxy",
	"read_service_ping",
}

func Token(c *gin.Context) {
	var tr TokenRequest
	params := map[string]string{}
	buf := new(bytes.Buffer)
	if err := c.ShouldBind(&tr); err != nil {
		Out(c, err.Error(), buf)
		return
	}
	s := sessions.Default(c)
	url := s.Get("GlUrl")
	token := s.Get("GlToken")
	git, err := gitlab.NewClient(token.(string), gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", url.(string))))
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
