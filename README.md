# Go/WebSocket + Redis/PubSub

This application starts a WebSocket Server in Go, default port 8080 (and 8443).
It will connect to the given Redis server and do PubSub things.

## Running the Server

    ./gwsrps --redis=127.0.0.1:6379,3
    ./gwsrps --redis=127.0.0.1:6379,3 --listen=8080 --listen-tls=8433 --tls-certificate=chain.pem

## Enable SSL/TLS

Please do this


## See Also

 * https://godoc.org/github.com/go-redis/redis
 * https://gist.github.com/miguelmota/ca4a7a66d8a7b014fad09d17df4b6e18
 * https://github.com/go-redis/redis
