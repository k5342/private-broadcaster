# private broadcaster

## Requirements
* Go
* Docker
* Twitter ConsumerKey / ConsumerSecret

## How to launch
Copy .env.example to .env and write your appropriate
environment variables into src/.env

```
mv srv/.env.example src/
vim src/.env
```

Install go packages listed below:
```
github.com/gin-contrib/sessions
github.com/gin-gonic/gin
github.com/jinzhu/gorm
github.com/jinzhu/gorm/dialects/sqlite
github.com/dghubble/oauth1
github.com/dghubble/oauth1/twitter
github.com/dghubble/go-twitter/twitter
github.com/joho/godotenv
```

Launch server following command:
```
cd src
go app.go
```

## How to deploy
compile .go files and docker image following command:

```shell
make
```

After build, run the container.
Bind port 8080 on host machines to port 8080 on container.

Now you can connect `localhost:8080` through your web browser.

```
docker run -p 8080:8080 private-broadcaster:0.5
```

## TODO
* Implement integration between nginx-rtmp-module and Web
  * Handle access control
* Appropriate nginx-rtmp-module settings
