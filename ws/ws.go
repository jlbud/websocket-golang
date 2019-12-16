package ws

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetCheckOrigin(f func(r *http.Request) bool) {
	upgrader.CheckOrigin = f
}

func NewUpgrader(f func(r *http.Request) bool) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: f,
	}
}

type HTTPContext struct {
	ResponseWriter http.ResponseWriter `json:"-"`
	Request        *http.Request       `json:"-"`
}

type WsIns struct {
	HTTPCtx     *HTTPContext
	ws          *websocket.Conn
	keepTimeout time.Duration
}

func NewWS(httpCtx *HTTPContext, h http.Header) (wsIns *WsIns, err error) {
	ws, err := upgrader.Upgrade(httpCtx.ResponseWriter, httpCtx.Request, h)
	if err != nil {
		return
	}
	wsIns = &WsIns{
		HTTPCtx:     httpCtx,
		keepTimeout: 60,
	}
	wsIns.ws = ws
	go wsIns.keep()

	return
}

func NewWSWithUpgrader(httpCtx *HTTPContext, h http.Header, upgrader websocket.Upgrader) (wsIns *WsIns, err error) {
	ws, err := upgrader.Upgrade(httpCtx.ResponseWriter, httpCtx.Request, h)
	if err != nil {
		return
	}
	wsIns = &WsIns{
		HTTPCtx:     httpCtx,
		keepTimeout: 60,
	}
	wsIns.ws = ws
	go wsIns.keep()

	return
}

func (wsIns *WsIns) Close() error {
	return wsIns.ws.Close()
}

func (wsIns *WsIns) keep() {
FOR:
	for {
		select {
		case <-context.Background().Done(): // TODO need to replace the context
			// 发送个信号给客户端，由客户端关闭
			err := wsIns.WriteCloseMessage(websocket.CloseServiceRestart, "keep ctx done")
			if err != nil {
				fmt.Println("keep ctx done:", err)
			}
			break FOR
		case <-time.After(wsIns.keepTimeout * time.Second):
			err := wsIns.WritePingMessage()
			if err != nil {
				fmt.Println("keep error: %v", err)
				fmt.Println(wsIns.Close())
				break FOR
			}
		}
	}
}
func (wsIns *WsIns) ReadMessage() (messageType int, p []byte, err error) {
	return wsIns.ws.ReadMessage()
}

func (wsIns *WsIns) WritePingMessage() (err error) {
	return wsIns.ws.WriteMessage(websocket.PingMessage, nil)
}

func (wsIns *WsIns) WriteTextMessage(data []byte) (err error) {
	return wsIns.ws.WriteMessage(websocket.TextMessage, data)
}

func (wsIns *WsIns) WriteCloseMessage(closeCode int, text string) error {
	return wsIns.ws.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, text))
}

func (wsIns *WsIns) IsWebSocketCloseError(err error) bool {
	if err == nil {
		return false
	}
	close := []int{
		// 正常关闭
		websocket.CloseNormalClosure,
		// 当客户端页面刷新，ReadMessage就报这个错误
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
		// 发送CloseMessage给客户端后，服务器端就会收到这个
		websocket.CloseNoStatusReceived,
	}
	// 服务器端或客户端close后，再对客户端写入，就会ErrCloseSent
	if websocket.IsCloseError(err, close...) || err == websocket.ErrCloseSent {
		fmt.Println("error: %v", err)
		return true
	}
	// 服务器端close后，再对客户端读取，就会这个错误
	if strings.Contains(err.Error(), "closed network") {
		fmt.Println("error: %v", err)
		return true
	}

	fmt.Println(err)
	return false
}
