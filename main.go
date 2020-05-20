package main

import 
(
	"github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
	"os"
	"github.com/markbates/goth/providers/google"
	"fmt"
	"log"
	"net/http"
	"html/template"
	"sort"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth"
	"github.com/gorilla/pat"
	"github.com/markbates/goth/providers/openidConnect"
)

func main() {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://18.191.255.130/auth/google/callback"),)

		openidConnect, _ := openidConnect.New(os.Getenv("OPENID_CONNECT_KEY"), os.Getenv("OPENID_CONNECT_SECRET"), "http://18.191.255.130/auth/openid-connect/callback", os.Getenv("OPENID_CONNECT_DISCOVERY_URL"))
		if openidConnect != nil {
			goth.UseProviders(openidConnect)
		}

		m := make(map[string]string)
		m["google"] = "Google"

		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

		p := pat.New()
		p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

			user, err := gothic.CompleteUserAuth(res, req)
			if err != nil {
				fmt.Fprintln(res, err)
				return
			}
			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(res, user)
		})

		p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
			gothic.Logout(res, req)
			res.Header().Set("Location", "/")
			res.WriteHeader(http.StatusTemporaryRedirect)
		})

		p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
			// try to get the user without re-authenticating
			if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
				t, _ := template.New("foo").Parse(userTemplate)
				t.Execute(res, gothUser)
			} else {
				gothic.BeginAuthHandler(res, req)
			}
		})

		p.Get("/", func(res http.ResponseWriter, req *http.Request) {
			t, _ := template.New("foo").Parse(indexTemplate)
			t.Execute(res, providerIndex)
		})

	log.Println("Websocket App start.")

    router := gin.Default()
	m1 := melody.New()


	rg := router.Group("/")
	//ログイン処理Oauthなし
	store := cookie.NewStore([]byte("secret"))
    router.Use(sessions.Sessions("mysession", store))
 
    router. POST("/login", func(c *gin.Context) {
        // セッションの作成
        session := sessions.Default(c)
        session.Set("loginUser", c.PostForm("userId"))
        session.Save()
        c.String(http.StatusOK, "ログイン完了")
    })
    rg.GET("/login", func(ctx *gin.Context) {
            http.ServeFile(ctx.Writer, ctx.Request, "login.html")
    })
    router. GET("/logout", func(c *gin.Context) {
        // セッションの破棄
        session := sessions.Default(c)
        session.Clear()
        session.Save()
        c.String(http.StatusOK, "ログアウトしました")
	})
	



    router.Static("/assets", "./assets")
    
    
    rg.GET("/", func(ctx *gin.Context) {
	    http.ServeFile(ctx.Writer, ctx.Request, "index.html")
    })

    rg.GET("/ws", func(ctx *gin.Context) {
        m1.HandleRequest(ctx.Writer, ctx.Request)
    })

    m1.HandleMessage(func(s *melody.Session, msg []byte) {
        m1.Broadcast(msg)
    })

    m1.HandleConnect(func(s *melody.Session) {
        log.Printf("websocket connection open. [session: %#v]\n", s)
    })

    m1.HandleDisconnect(func(s *melody.Session) {
        log.Printf("websocket connection close. [session: %#v]\n", s)
    })

    // Listen and server on 0.0.0.0:8989
    router.Run(":8080")

    fmt.Println("Websocket App End.")

}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`






    





    


