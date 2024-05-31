package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

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
	if err := Output(in).Render(context.Background(), b); err != nil {
		c.String(200, err.Error())
	}
	c.String(200, b.String())
}

func OutToken(c *gin.Context, in string, b *bytes.Buffer) {
	if err := OutputToken(in).Render(context.Background(), b); err != nil {
		c.String(200, err.Error())
	}
	c.String(200, b.String())
}

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatalln("DB init error", err)
	}
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	year, _, _ := time.Now().Date()

	// index
	r.GET("/", gin.WrapH(templ.Handler(Site(year))))

	// htmx html returns
	r.GET("/home", gin.WrapH(templ.Handler(Home())))
	r.GET("/serviceaccount", AuthMiddleware(), gin.WrapH(templ.Handler(CreateServiceUser())))
	r.GET("/token", AuthMiddleware(), gin.WrapH(templ.Handler(CreatePersonalAccessToken(Scopes))))
	r.GET("/login", gin.WrapH(templ.Handler(Auth("login"))))
	r.GET("/logout", AuthMiddleware(), gin.WrapH(templ.Handler(LogoutComponent())))
	r.GET("/register", gin.WrapH(templ.Handler(Auth("register"))))
	r.GET("/sessionstore", AuthMiddleware(), gin.WrapH(templ.Handler(SessionComponent())))
	r.GET("/nav/login", gin.WrapH(templ.Handler(NavbarLoggedIn())))
	r.GET("/nav/logout", gin.WrapH(templ.Handler(NavbarLoggedOut())))

	// login + register + session
	r.POST("/login", Login(db))
	r.POST("/logout", Logout())
	r.POST("/register", Register(db))

	// buisness logic
	r.POST("/serviceaccount", Serviceaccount)
	r.POST("/token", Token)
	r.POST("/sessionstore", SessionStore)

	if err := r.Run(":8080"); err != nil {
		log.Fatalln("Gin Run Error: ", err)
	}
}
