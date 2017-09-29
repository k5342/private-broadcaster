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
	. "private-broadcaster/models"
)

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
		c.JSON(200, gin.H{
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
	})
	
	r.GET("/live/:screen_name", func(c *gin.Context) {
//		screen_name := c.Param("screen_name")
	})
	
	r.Run()
}
