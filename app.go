package main

import (
	"net/http"
	"time"
	"os"
	"log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/dghubble/oauth1"
	twauth "github.com/dghubble/oauth1/twitter"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/joho/godotenv"
	"github.com/thanhpk/randstr"
	. "private-broadcaster/models"
)

func GetCurrentUser(db *gorm.DB, session sessions.Session) User {
	var twitter_id = session.Get("twitter_id").(int64)
	var user User
	
	db.Where(User{TwitterID: twitter_id}).First(&user)
	
	return user
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	db, _ := gorm.Open("sqlite3", "./database.db")
	db.AutoMigrate(&User{}, &Broadcast{})
	defer db.Close()
	
	twitter_config := oauth1.Config {
		ConsumerKey:	os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret:	os.Getenv("TWITTER_CONSUMER_SECRET"),
		CallbackURL:	"http://localhost:8080/login/callback",
		Endpoint:		twauth.AuthorizeEndpoint,
	}
	
	r := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("login", store))
	
	r.LoadHTMLGlob("templates/*")
	
	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		
		c.HTML(http.StatusOK, "index.html", gin.H{
			"is_login": session.Get("is_login"),
			"screen_name": session.Get("screen_name"),
		})
	})
	
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	
	r.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		
		requestToken, requestSecret, err := twitter_config.RequestToken()
		session.Set("request_token", requestToken)
		session.Set("request_secret", requestSecret)
		session.Save()
		
		if err != nil {
			log.Fatal(err)
		}
		
		authorizationURL, err := twitter_config.AuthorizationURL(requestToken)
		
		if err != nil {
			log.Fatal(err)
		}
		
		log.Print(authorizationURL.String())
		c.Redirect(http.StatusFound, authorizationURL.String())
	})
	
	r.GET("/login/callback", func(c *gin.Context) {
		session:= sessions.Default(c)
		
		requestToken := session.Get("request_token").(string)
		requestSecret := session.Get("request_secret").(string)
		verifier := c.Query("oauth_verifier")
		
		accessToken, accessSecret, err := twitter_config.AccessToken(requestToken, requestSecret, verifier)
		
		if err == nil {
			session.Set("access_token", accessToken)
			session.Set("access_secret", accessSecret)
			session.Save()
			
			token := oauth1.NewToken(accessToken, accessSecret)
			httpClient := twitter_config.Client(oauth1.NoContext, token)
			client := twitter.NewClient(httpClient)
			user, _, err := client.Accounts.VerifyCredentials(nil)
			
			if err != nil {
			}
			
			session.Set("twitter_id", user.ID)
			session.Set("screen_name", user.ScreenName)
			session.Set("name", user.Name)
			session.Set("is_login", true)
			session.Save()
			
			u := User {
				ScreenName:		user.ScreenName,
				Name:			user.Name,
				LastLoginedAt:	time.Now(),
			}
			
			db.Where(User{TwitterID: user.ID}).FirstOrCreate(&u)
			
			c.Redirect(http.StatusFound, "/")
		} else {
			log.Fatal(err)
		}
	})
	
	r.GET("/logout", func(c *gin.Context) {
		session:= sessions.Default(c)
		session.Clear()
		session.Set("is_login", false)
		session.Save()
		
		c.Redirect(http.StatusFound, "/")
	})
	
	r.GET("/start", func(c *gin.Context) {
//		session := sessions.Default(c)
		
		c.HTML(http.StatusOK, "start.html", gin.H{
			"csrf_token": "stub",
		})
	})

	r.POST("/create", func(c *gin.Context) {
		session:= sessions.Default(c)
		
		u := GetCurrentUser(db, session)
		
		var b Broadcast
		err := db.Model(&u).
					Not(Broadcast{EndedAt: nil}).
					Order("created_at desc").
					First(&b).
					Error
		
		if err == nil {
			// already started broadcast
			c.Redirect(http.StatusFound, "/create/done?reason=duplicate")
		}
		
		err = db.Create(&Broadcast{
			User:		u,
			RTMPName:	randstr.RandomHex(16),
			Password:	c.PostForm("password"),
		}).Error
		
		if err != nil {
			log.Fatal(err)
		}
		
		c.Redirect(http.StatusFound, "/create/done")
	})
	
	r.GET("/create/done", func(c *gin.Context) {
		session:= sessions.Default(c)
		
		u := GetCurrentUser(db, session)
		
		var b Broadcast
		err := db.Model(&u).
					Not(Broadcast{EndedAt: nil}).
					Order("created_at desc").
					First(&b).
					Error
		
		if err == nil {
			c.HTML(http.StatusOK, "done.html", gin.H{
				"host": "b.ksswre.net",
				"rtmp_endpoint": "rtmp://b.ksswre.net/hls",
				"screen_name": u.ScreenName,
				"stream_key": b.RTMPName,
			})
		} else {
			log.Print(err)
			c.Redirect(http.StatusFound, "/")
		}
	})
	
	
	r.GET("/live/:screen_name", func(c *gin.Context) {
//		screen_name := c.Param("screen_name")
	})

	r.POST("/api/on_publish", func(c *gin.Context) {
		var name = c.PostForm("name")
		
		var b Broadcast
		err := db.Where(Broadcast{RTMPName: name}).First(&b).Error
		
		if err != nil {
			log.Print(err)
			c.String(403, "Access not allowed.")
			return
		}
		
		c.String(200, "OK.")
	})
	
	r.Run()
}
