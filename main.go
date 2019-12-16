package main

import (
	"fmt"
	"net/http"

	ws2 "github/websocket-golang/ws"
)

func main() {
	http.HandleFunc("/ws/chat", WsChat)
	http.ListenAndServe(":8080", nil)
}

func WsChat(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	fmt.Printf("WsChat key: %s", key)
	httpCtx := &ws2.HTTPContext{
		ResponseWriter: w,
		Request:        r,
	}
	ws, err := ws2.NewWS(httpCtx, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, b, err := ws.ReadMessage()
		if err != nil {
			ws.Close()
			return
		}
		fmt.Println(fmt.Sprintf("WsChart receive message: %s", string(b)))
		ws.WriteTextMessage(b)
	}
}
