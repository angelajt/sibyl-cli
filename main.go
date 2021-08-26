package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/sacOO7/gowebsocket"
	"github.com/synacor/sibyl/server"
)

func main() {

	topic := flag.String("t", "", "new room topic sentence")
	room := flag.String("r", "", "room name")
	deck := flag.String("d", "", "deck name")
	flag.Parse()

	if len(*room) == 0 {
		log.Fatal("no room given")
	}

	hostport := os.Getenv("SIBYL_HOST")

	if len(hostport) == 0 {
		log.Fatal("no host given. set SIBYL_HOST to host:port")
	}

	geturl := fmt.Sprintf("http://%s/r/%s", hostport, *room)
	resp, err := http.Get(geturl)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// extract token from html

	re := regexp.MustCompile("Token:\\s+\"(.+?)\"")
	token := re.FindStringSubmatch(string(body))[1]

	if len(*topic) != 0 {
		setTopic(hostport, *room, token, *topic)
	}

	if len(*deck) != 0 {
		setDeck(hostport, *room, token, *deck)
	}
}

func setDeck(hostport, room, token, deck string) {
	req := server.WsRequest{
		Room:   room,
		Action: server.WsRequestActionDeck,
		Deck:   deck,
		Token:  token,
	}
	set(req, hostport)
}

func setTopic(hostport, room, token, topic string) {
	req := server.WsRequest{
		Room:   room,
		Action: server.WsRequestActionTopic,
		Value:  topic,
		Token:  token,
	}
	set(req, hostport)
}

func set(req server.WsRequest, hostport string) {

	buf, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	txt := string(buf)

	wsurl := fmt.Sprintf("ws://%s/ws", hostport)
	username := "sibylcli"
	url := fmt.Sprintf("%s?room=%s&token=%s&username=%s", wsurl, req.Room, url.QueryEscape(req.Token), username)
	socket := gowebsocket.New(url)

	socket.Connect()
	defer socket.Close()
	socket.SendText(txt)

	// prevent connection from closing before server processes message
	// XXX instead of sleeping, we should wait for the server to send a message before we return
	time.Sleep(3 * time.Second)
}

/* in case we ever want to create rooms

data := url.Values{
		"room": {room},
		"deck": {"hewwo"},
	}

	resp, err := http.PostForm("http://%s/create", hostport, data)
	if err != nil {
		log.Fatal(err)
	}

*/
