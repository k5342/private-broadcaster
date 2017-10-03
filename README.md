# private broadcaster

## Requirements
* Go
* Docker (requires 17.05 later)
* Twitter ConsumerKey / ConsumerSecret
* [glide](https://glide.sh/); for package management

## How to launch web app
Copy `.env.example` to `.env` and write your appropriate
environment variables into `.env`

```
cp .env.example .env
vim .env
```

Install go packages using glide:
```
glide install
glide rebuild
```

Launch server following command:
```
go run app.go
```

Now, you can connect through your web browser `localhost:8080`

## How to launch web app and nginx with *nginx-rtmp-module*
compile `.go` files and docker image following command:

```shell
./build
```

After build, run the container.
Bind port 8080 on host machines to port 8080 on container.

Now you can connect `localhost:8080` through your web browser.

```
./run
```

## TODO
* Implement integration between nginx-rtmp-module and web app
  * Handle access control
* Appropriate nginx-rtmp-module settings
