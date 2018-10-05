# WebMusicPlayerServer

This is the backend server for web music player: [KiwiPlayer](https://pinkiebala.nctu.me/KiwiPlayer/)

# Setup

Dependency. Please install following packages

    go get github.com/BurntSushi/toml
    go get github.com/gin-contrib/cors
    go get github.com/gin-gonic/gin
    go get gopkg.in/mgo.v2
    go get gopkg.in/mgo.v2/bson
    go get github.com/fsnotify/fsnotify
    
Next, create file `config.toml` by copy from `config-example.toml`.

Yon can now configure for your music server.

We use mongodb database. Make sure you set it up.

# Run
You can direct run it by `go run *.go` or build it with `go build`
