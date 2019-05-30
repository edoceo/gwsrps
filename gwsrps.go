/**
 * Go WebSocket + Redis/PubSub
 */

package main

//import "os"
import "flag"
import "fmt"
import "net/http"

import "github.com/gorilla/websocket"


type WS_Client struct {
	id string
	ws *websocket.Conn
}

/*
var (
    gStore      *Store
    gPubSubConn *redis.PubSubConn
    gRedisConn  = func() (redis.Conn, error) { 
        return redis.Dial("tcp", ":6379") 
    }
)
func init() {
    gStore = &Store{
        Users: make([]*User, 0, 1),
    }
}
*/


func main() {

	//cli_args := os.Args[1:]
	//fmt.Println(cli_args)

	redis := flag.String("redis", "127.0.0.1:6357", "Redis Server host:port:db")
	tcpListen := flag.String("listen", ":8080", "Listen Address for TCP")
	tlsListen := flag.String("listen-tls", ":8443", "Listen Address for TLS/SSL")
	tlsFile   := flag.String("tls-stuff", "", "Stuff")

	flag.Parse()


	fmt.Println("Connecting to Redis:", *redis)
	fmt.Println("TCP:", *tcpListen)
	fmt.Println("TLS:", *tlsListen)
	fmt.Println("CRT:", *tlsFile)

	// Fork to Background

	// http.HandleFunc("/"
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("HTTP Connection")
	})

	err := http.ListenAndServe(*tcpListen, nil)
	if err != nil {
		panic(err)
	}


}
