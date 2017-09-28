# private broadcaster

## Requirements
* Go
* Docker

## How to use
compile .go files and docker image following command:

```shell
make
```

After build, run the container.
Bind port 8080 on host machines to port 8080 on container.

You can connect `localhost:8080` through your web browser.

```
docker run -p 8080:8080 private-broadcaster:0.5
```

## TODO
* Implement integration between nginx-rtmp-module and Web
  * Handle access control
* Appropriate nginx-rtmp-module settings
