package main

import (
	"bytes"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/xanzy/go-gitlab"
)

type ServiceAccountRequest struct {
	Name        string `form:"saName"`
	Displayname string `form:"saDisplayName"`
}

func Serviceaccount(c *gin.Context) {
	var sar ServiceAccountRequest
	params := map[string]string{}
	buf := new(bytes.Buffer)
	if err := c.ShouldBind(&sar); err != nil {
		Out(c, err.Error(), buf)
		return
	}
	s := sessions.Default(c)
	url := s.Get("GlUrl")
	token := s.Get("GlToken")
	if url == nil || token == nil {
		Out(c, "GitlabUrl or GitlabToken not set!", buf)
		return
	}
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
