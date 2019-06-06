# Go/WebSocket + Redis/PubSub

This application starts a WebSocket Server in Go, default port 8080 (and 8443).
It will connect to the given Redis server and do PubSub things.


## Building

	go get github.com/gorilla/websocket
	go get github.com/go-redis/redis
    go get github.com/oklog/ulid
    go build


## Running the Server

Here what all the options look like, and shows their default values.

    ./gwsrps \
    	--redis=127.0.0.1:6379:0 \
    	--tcp-addr=8080 \
    	--tls-addr=8433 \
    	--tls-pem=server.pem \
    	--tls-key=server.key \
    	--path=/ \
    	--origin='*'


* **--redis**=$HOST:$POST:$DATABASE_ID
* **--tcp-addr**=8080 -- the listen address for TCP/HTTP/WS connections
* **--tls-addr**=8443 -- the listen address for SSL/TLS/HTTPS/WSS connections
* **--tls-pem** and **--tls-key** are the certificate/chain PEM files and the key.
* **--path**=/ -- the websocket path
* **--origin='*'** -- the allowed origin glob-like pattern


## Enable SSL/TLS

Please do this but use proper SSL Certificates.
One can rename the files, just pass proper options

    openssl genrsa -out server.key 2048
    openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650


## See Also

 * https://github.com/gorilla/websocket
 * https://github.com/oklog/ulid/
 * https://github.com/go-redis/redis

 * https://hackernoon.com/communicating-go-applications-through-redis-pub-sub-messaging-paradigm-df7317897b13
