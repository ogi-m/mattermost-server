// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	SOCKET_MAX_MESSAGE_SIZE_KB  = 8 * 1024 // 8KB
	PING_TIMEOUT_BUFFER_SECONDS = 5
)

type msgType int

const (
	msgTypeJSON msgType = iota + 1
	msgTypePong
)

type writeMessage struct {
	msgType  msgType
	contents interface{}
}

// WebSocketClient stores the necessary information required to
// communicate with a WebSocket endpoint.
type WebSocketClient struct {
	Url                string                  // The location of the server like "ws://localhost:8065"
	ApiUrl             string                  // The API location of the server like "ws://localhost:8065/api/v3"
	ConnectUrl         string                  // The WebSocket URL to connect to like "ws://localhost:8065/api/v3/path/to/websocket"
	Conn               *websocket.Conn         // The WebSocket connection
	AuthToken          string                  // The token used to open the WebSocket connection
	Sequence           int64                   // The ever-incrementing sequence attached to each WebSocket action
	PingTimeoutChannel chan bool               // The channel used to signal ping timeouts
	EventChannel       chan *WebSocketEvent    // The channel used to receive various events pushed from the server. For example: typing, posted
	ResponseChannel    chan *WebSocketResponse // The channel used to receive responses for requests made to the server
	ListenError        *AppError               // A field that is set if there was an abnormal closure of the WebSocket connection
	writeChan          chan writeMessage

	pingTimeoutTimer  *time.Timer
	closePingWatchdog chan struct{}

	quitWriterChan chan struct{}
	quitReaderChan chan struct{}
	closeOnce      sync.Once
	closed         bool
}

// NewWebSocketClient constructs a new WebSocket client with convenience
// methods for talking to the server.
func NewWebSocketClient(url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClientWithDialer(websocket.DefaultDialer, url, authToken)
}

// NewWebSocketClientWithDialer constructs a new WebSocket client with convenience
// methods for talking to the server using a custom dialer.
func NewWebSocketClientWithDialer(dialer *websocket.Dialer, url, authToken string) (*WebSocketClient, *AppError) {
	conn, _, err := dialer.Dial(url+API_URL_SUFFIX+"/websocket", nil)
	if err != nil {
		return nil, NewAppError("NewWebSocketClient", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	client := &WebSocketClient{
		Url:                url,
		ApiUrl:             url + API_URL_SUFFIX,
		ConnectUrl:         url + API_URL_SUFFIX + "/websocket",
		Conn:               conn,
		AuthToken:          authToken,
		Sequence:           1,
		PingTimeoutChannel: make(chan bool, 1),
		EventChannel:       make(chan *WebSocketEvent, 100),
		ResponseChannel:    make(chan *WebSocketResponse, 100),
		writeChan:          make(chan writeMessage),
		closePingWatchdog:  make(chan struct{}),
		quitWriterChan:     make(chan struct{}),
		quitReaderChan:     make(chan struct{}),
	}

	client.configurePingHandling()
	go client.writer()

	client.SendMessage(WEBSOCKET_AUTHENTICATION_CHALLENGE, map[string]interface{}{"token": authToken})

	return client, nil
}

// NewWebSocketClient4 constructs a new WebSocket client with convenience
// methods for talking to the server. Uses the v4 endpoint.
func NewWebSocketClient4(url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClient4WithDialer(websocket.DefaultDialer, url, authToken)
}

// NewWebSocketClient4WithDialer constructs a new WebSocket client with convenience
// methods for talking to the server using a custom dialer. Uses the v4 endpoint.
func NewWebSocketClient4WithDialer(dialer *websocket.Dialer, url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClientWithDialer(dialer, url, authToken)
}

// Connect creates a websocket connection with the given ConnectUrl.
// This is racy and error-prone should not be used. Use any of the New* functions to create a websocket.
func (wsc *WebSocketClient) Connect() *AppError {
	return wsc.ConnectWithDialer(websocket.DefaultDialer)
}

// ConnectWithDialer creates a websocket connection with the given ConnectUrl using the dialer.
// This is racy and error-prone and should not be used. Use any of the New* functions to create a websocket.
func (wsc *WebSocketClient) ConnectWithDialer(dialer *websocket.Dialer) *AppError {
	var err error
	wsc.Conn, _, err = dialer.Dial(wsc.ConnectUrl, nil)
	if err != nil {
		return NewAppError("Connect", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	// Super racy and should not be done anyways.
	// All of this needs to be redesigned for v6.
	if wsc.writeChan == nil || wsc.closed {
		wsc.writeChan = make(chan writeMessage)
		go wsc.writer()
	}
	if wsc.closePingWatchdog == nil || wsc.closed {
		wsc.closePingWatchdog = make(chan struct{})
		wsc.configurePingHandling()
	}
	if wsc.quitWriterChan == nil || wsc.closed {
		wsc.quitWriterChan = make(chan struct{})
	}
	if wsc.quitReaderChan == nil || wsc.closed {
		wsc.quitReaderChan = make(chan struct{})
	}
	if wsc.closed {
		wsc.closeOnce = sync.Once{}
		wsc.closed = false
	}

	wsc.EventChannel = make(chan *WebSocketEvent, 100)
	wsc.ResponseChannel = make(chan *WebSocketResponse, 100)

	wsc.SendMessage(WEBSOCKET_AUTHENTICATION_CHALLENGE, map[string]interface{}{"token": wsc.AuthToken})

	return nil
}

// Close will close the underlying websocket connection and clean up
// all the resources related to the WebSocketClient. The client should
// not be used after it has been closed.
func (wsc *WebSocketClient) Close() {
	// We use a closeOnce to guarantee that if a Close happens due to ReadMessage error,
	// then clients calling Close after that makes this a noop.
	// Essentially, there are 2 ways to shut down the client.
	// 1. Call the Close method.
	// 2. If an error occurs during ReadMessage.
	wsc.closeOnce.Do(func() {
		// close writer
		close(wsc.quitWriterChan)
		close(wsc.writeChan)

		// close connection
		wsc.Conn.Close()

		// close listener
		// If the listener returns from a ReadMessage error,
		// then this signal will be needlessly sent, but it's a good thing
		// to close a channel on our way out.
		close(wsc.quitReaderChan)

		// close everything else
		close(wsc.EventChannel)
		close(wsc.ResponseChannel)
		close(wsc.closePingWatchdog)

		wsc.closed = true
	})
}

func (wsc *WebSocketClient) writer() {
	for {
		select {
		case msg := <-wsc.writeChan:
			switch msg.msgType {
			case msgTypeJSON:
				wsc.Conn.WriteJSON(msg.contents)
			case msgTypePong:
				wsc.Conn.WriteMessage(websocket.PongMessage, []byte{})
			}
		case <-wsc.quitWriterChan:
			return
		}
	}
}

func (wsc *WebSocketClient) Listen() {
	go func() {
		defer wsc.Close()

		for {
			select {
			case <-wsc.quitReaderChan:
				return
			default:
				var rawMsg json.RawMessage
				var err error
				if _, rawMsg, err = wsc.Conn.ReadMessage(); err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
						wsc.ListenError = NewAppError("NewWebSocketClient", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
					}

					return
				}

				event := WebSocketEventFromJson(bytes.NewReader(rawMsg))
				if event == nil {
					continue
				}
				if event.IsValid() {
					wsc.EventChannel <- event
					continue
				}

				var response WebSocketResponse
				if err := json.Unmarshal(rawMsg, &response); err == nil && response.IsValid() {
					wsc.ResponseChannel <- &response
					continue
				}
			}
		}
	}()
}

func (wsc *WebSocketClient) SendMessage(action string, data map[string]interface{}) {
	req := &WebSocketRequest{}
	req.Seq = wsc.Sequence
	req.Action = action
	req.Data = data

	wsc.Sequence++
	wsc.writeChan <- writeMessage{
		msgType:  msgTypeJSON,
		contents: req,
	}
}

// UserTyping will push a user_typing event out to all connected users
// who are in the specified channel
func (wsc *WebSocketClient) UserTyping(channelId, parentId string) {
	data := map[string]interface{}{
		"channel_id": channelId,
		"parent_id":  parentId,
	}

	wsc.SendMessage("user_typing", data)
}

// GetStatuses will return a map of string statuses using user id as the key
func (wsc *WebSocketClient) GetStatuses() {
	wsc.SendMessage("get_statuses", nil)
}

// GetStatusesByIds will fetch certain user statuses based on ids and return
// a map of string statuses using user id as the key
func (wsc *WebSocketClient) GetStatusesByIds(userIds []string) {
	data := map[string]interface{}{
		"user_ids": userIds,
	}
	wsc.SendMessage("get_statuses_by_ids", data)
}

func (wsc *WebSocketClient) configurePingHandling() {
	wsc.Conn.SetPingHandler(wsc.pingHandler)
	wsc.pingTimeoutTimer = time.NewTimer(time.Second * (60 + PING_TIMEOUT_BUFFER_SECONDS))
	go wsc.pingWatchdog()
}

func (wsc *WebSocketClient) pingHandler(appData string) error {
	if !wsc.pingTimeoutTimer.Stop() {
		<-wsc.pingTimeoutTimer.C
	}

	wsc.pingTimeoutTimer.Reset(time.Second * (60 + PING_TIMEOUT_BUFFER_SECONDS))
	wsc.writeChan <- writeMessage{
		msgType: msgTypePong,
	}
	return nil
}

func (wsc *WebSocketClient) pingWatchdog() {
	select {
	case <-wsc.pingTimeoutTimer.C:
		wsc.PingTimeoutChannel <- true
	case <-wsc.closePingWatchdog:
	}
}
