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

var twitter_config oauth1.Config

func GetCurrentUser(db *gorm.DB, session sessions.Session) User {
	var twitter_id = session.Get("twitter_id").(int64)
	var user User
	db.Where(User{TwitterID: twitter_id}).First(&user)
	return user
}

func GetUserByScreenName(db *gorm.DB, screen_name string) User {
	var user User
	db.Where(User{ScreenName: screen_name}).Order("updated_at DESC").First(&user)
	return user
}

func GetLatestBroadcastByScreenName(db *gorm.DB, screen_name string) Broadcast {
	var broadcast Broadcast
	var u = GetUserByScreenName(db, screen_name)
	db.Preload("User").Model(&u).Order("created_at DESC").First(&broadcast)
	return broadcast
}

func CheckCanPlay(db *gorm.DB, session sessions.Session, broadcast_user_screen_name string) bool {
	u := GetCurrentUser(db, session)
	b := GetLatestBroadcastByScreenName(db, broadcast_user_screen_name)
	
	var ba BroadcastApproval
	err := db.Debug().Where(&BroadcastApproval{BroadcastID: b.ID, UserID: u.ID}).Find(&ba).Error
	
	if err == nil {
		return true
	} else {
		return false
	}
}

func TwitterClient(session sessions.Session) *twitter.Client {
	accessToken := session.Get("access_token").(string)
	accessSecret := session.Get("access_secret").(string)
	
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := twitter_config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}

func AuthorizeBroadcast(db *gorm.DB, session sessions.Session, broadcast Broadcast) {
	u := GetCurrentUser(db, session)
	
	db.Create(&BroadcastApproval{
		Broadcast: broadcast,
		User: u,
	})
}

type CurrentUser struct {
	Test int
	PlayableBroadcasts map[uint]bool
}

func main() {
	err := godotenv.Load()
	
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	twitter_config = oauth1.Config {
		ConsumerKey:	os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret:	os.Getenv("TWITTER_CONSUMER_SECRET"),
		CallbackURL:	"http://localhost:8080/login/callback",
		Endpoint:		twauth.AuthorizeEndpoint,
	}
	
	db, _ := gorm.Open("sqlite3", "./database.db")
	db.AutoMigrate(&User{}, &Broadcast{}, &BroadcastApproval{})
	defer db.Close()
	
	r := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("login", store))
	
	r.LoadHTMLGlob("templates/*")
	
	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		
		cu := session.Get("current_user")
		
		if cu != nil {
			log.Printf("#g", cu.(CurrentUser))
		}
			
		
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
		
		log.Printf("%#v", twitter_config)
		requestToken, requestSecret, err := twitter_config.RequestToken()
		
		if err != nil {
			log.Fatal(err)
		}
		
		session.Set("request_token", requestToken)
		session.Set("request_secret", requestSecret)
		session.Save()
		
		authorizationURL, err := twitter_config.AuthorizationURL(requestToken)
		
		if err != nil {
			log.Print(2)
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
			log.Printf("%#v", session)
			session.Save()
			
			user, _, err := TwitterClient(session).Accounts.VerifyCredentials(nil)
			
			if err != nil {
			}
			
			session.Set("twitter_id", user.ID)
			session.Set("screen_name", user.ScreenName)
			session.Set("name", user.Name)
			session.Set("is_login", true)
			session.Save()
			
			session.Set("current_user", CurrentUser{
				PlayableBroadcasts: make(map[uint]bool),
			})
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
					Related(&b).
					Not(Broadcast{EndedAt: nil}).
					Order("created_at desc").
					Limit(1).
					Error
		
		if err == nil {
			// already started broadcast
			c.Redirect(http.StatusFound, "/create/done?reason=duplicate")
			return
		} else {
			log.Print(err)
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
					Related(&b).
					Not(Broadcast{EndedAt: nil}).
					Order("created_at desc").
					Limit(1).
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
		screen_name := c.Param("screen_name")
		session:= sessions.Default(c)
		if CheckCanPlay(db, session, screen_name) {
			c.HTML(http.StatusOK, "play.html", gin.H{
				"is_login": session.Get("is_login"),
				"screen_name": session.Get("screen_name"),
			})
		} else {
			c.Redirect(http.StatusFound, "/live/" + screen_name + "/auth")
		}
	})
	
	r.GET("/live/:screen_name/auth", func(c *gin.Context) {
		screen_name := c.Param("screen_name")
		session:= sessions.Default(c)
		
		if CheckCanPlay(db, session, screen_name) {
			c.Redirect(http.StatusFound, "./")
			return
		} else {
			b := GetLatestBroadcastByScreenName(db, screen_name)
			
			// try auth by twitter
			rs, _, err := TwitterClient(session).Friendships.Show(&twitter.FriendshipShowParams{
				SourceID: GetCurrentUser(db, session).TwitterID,
				TargetID: b.User.TwitterID,
			})
			
			if err != nil {
				log.Print(err)
			}
			
			if rs.Source.CanDM == true {
				AuthorizeBroadcast(db, session, b)
				c.Redirect(http.StatusFound, "./")
				return
			}
			
			// otherwise use password auth
			if b.Password == "" {
				c.String(403, "You are not allowed to play this broadcast.")
				return
			} else {
				c.HTML(http.StatusOK, "auth.html", gin.H{
					"broadcast": b,
				})
				return
			}
		}
	})
	
	r.GET("/video/:screen_name", func(c *gin.Context) {
		screen_name := c.Param("screen_name")
		session:= sessions.Default(c)
		
		if CheckCanPlay(db, session, screen_name) {
			// do something
		} else {
			c.AbortWithStatus(403)
		}
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
