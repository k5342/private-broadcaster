# private broadcaster

## Requirements
* Go
* Docker
* Twitter ConsumerKey / ConsumerSecret
* [glide](https://glide.sh/); for package management

## How to launch
Copy .env.example to .env and write your appropriate
environment variables into .env

```
cp .env.example .env
vim .env
```

Install go packages using glide:
```
glide install
```

Launch server following command:
```
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
