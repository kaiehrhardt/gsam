package main

import (
	"bytes"
	"database/sql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u User
		buf := new(bytes.Buffer)
		if err := c.ShouldBind(&u); err != nil {
			Out(c, err.Error(), buf)
			return
		}
		if err := u.InsertDb(db); err != nil {
			Out(c, err.Error(), buf)
			return
		}
	}
}

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u User
		buf := new(bytes.Buffer)
		if err := c.ShouldBind(&u); err != nil {
			Out(c, err.Error(), buf)
			return
		}
		ok := u.VerifyLogin(db)
		if ok {
			session := sessions.Default(c)
			session.Set("user", u.Name)
			if err := session.Save(); err != nil {
				Out(c, err.Error(), buf)
				return
			}
			c.Header("HX-Trigger", "logged-in")
			Out(c, "Successfully logged in!", buf)
		} else {
			Out(c, "You are not registered, yet!", buf)
		}
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf := new(bytes.Buffer)
		session := sessions.Default(c)
		user := session.Get("user")
		if user != nil {
			session.Delete("user")
			session.Delete("GlUrl")
			session.Delete("GlToken")
			if err := session.Save(); err != nil {
				Out(c, err.Error(), buf)
				return
			}
			c.Header("HX-Trigger", "logged-out")
			Out(c, "Successfully logged out!", buf)
		} else {
			Out(c, "You are not logged in, yet!", buf)
		}
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf := new(bytes.Buffer)
		session := sessions.Default(c)
		user := session.Get("user")

		if user != nil {
			c.Next()
		} else {
			Out(c, "Please login!", buf)
			c.Abort()
		}
	}
}
