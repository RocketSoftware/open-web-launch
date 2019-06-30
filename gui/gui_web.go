// +build web

package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

type webSocketMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
}

type WebGUI struct {
	port             int
	address          string
	url              string
	server           *http.Server
	webSocketChannel chan string
}

func NewWebGUI() GUI {
	return GUI(&WebGUI{})
}

func (gui *WebGUI) Start(windowTitle string) error {
	const port = 18485
	address := fmt.Sprintf("localhost:%d", port)
	url := "http://" + address
	server := &http.Server{Addr: address}
	gui.server = server
	statusServer := func(ws *websocket.Conn) {
		for message := range gui.webSocketChannel {
			ws.Write([]byte(message))
		}
	}
	http.Handle("/status", websocket.Handler(statusServer))
	// http.Handle("/", http.FileServer(http.Dir("assets")))
	http.Handle("/", http.FileServer(assetFS()))
	ch := make(chan struct{})
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	go func() {
		ch <- struct{}{}
		err := server.Serve(listener)
		if err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()
	<-ch
	gui.address = address
	gui.port = port
	gui.url = url
	gui.server = server
	gui.webSocketChannel = make(chan string)
	return browser.OpenURL(url)
}

func (gui *WebGUI) Terminate() error {
	if gui.server != nil {
		close(gui.webSocketChannel)
		return gui.server.Shutdown(context.Background())
	}
	return errors.New("server is not running")
}

func (gui *WebGUI) SendTextMessage(message string) error {
	return gui.sendMessage("message", message)
}

func (gui *WebGUI) SendErrorMessage(err error) error {
	return gui.sendMessage("error", err.Error())
}

func (gui *WebGUI) SendCloseMessage() error {
	return gui.sendMessage("close", "")
}

func (gui *WebGUI) SetTitle(title string) error {
	return gui.sendMessage("title", title)
}

func (gui *WebGUI) sendMessage(messageType string, payload string) error {
	message := webSocketMessage{messageType, payload}
	data, err := json.Marshal(&message)
	if err != nil {
		return errors.Wrapf(err, "unable to serialize Web Socket message %v\n", message)
	}
	select {
	case gui.webSocketChannel <- string(data):
	}
	return nil
}

func (gui *WebGUI) SetProgressMax(val int) {
	gui.sendMessage("progress_max", strconv.Itoa(val))
}

func (gui *WebGUI) ProgressStep() {
	gui.sendMessage("progress_step", "")
}

func (gui *WebGUI) Wait() {
}

func (gui *WebGUI) Closed() bool {
	return false
}
