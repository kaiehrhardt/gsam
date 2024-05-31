package main

import (
	"bytes"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type GitlabConnection struct {
	GlUrl   string `form:"glUrl" binding:"required"`
	GlToken string `form:"glToken" binding:"required"`
}

func SessionStore(c *gin.Context) {
	var gc GitlabConnection
	buf := new(bytes.Buffer)

	if err := c.ShouldBind(&gc); err != nil {
		Out(c, err.Error(), buf)
		return
	}

	session := sessions.Default(c)
	url := session.Get("GlUrl")
	token := session.Get("GlToken")
	if url == nil || token == nil {
		session.Set("GlUrl", gc.GlUrl)
		session.Set("GlToken", gc.GlToken)
		if err := session.Save(); err != nil {
			Out(c, err.Error(), buf)
			return
		}
		Out(c, "GitlabUrl and GlToken successfully set.", buf)
	} else {
		Out(c, "GitlabUrl and GlToken already set.", buf)
	}
}
