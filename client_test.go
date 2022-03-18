package centrifuge

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type testEventHandler struct {
	onConnect    func(ConnectEvent)
	onDisconnect func(DisconnectEvent)
	onError      func(ErrorEvent)
}

func (h *testEventHandler) OnConnect(e ConnectEvent) {
	if h.onConnect != nil {
		h.onConnect(e)
	}
}

func (h *testEventHandler) OnDisconnect(e DisconnectEvent) {
	if h.onDisconnect != nil {
		h.onDisconnect(e)
	}
}

func (h *testEventHandler) OnError(e ErrorEvent) {
	if h.onError != nil {
		h.onError(e)
	}
}

type testSubscriptionHandler struct {
	onSubscribe   func(SubscribeEvent)
	onError       func(SubscriptionErrorEvent)
	onPublication func(PublicationEvent)
	onUnsubscribe func(UnsubscribeEvent)
}

func (h *testSubscriptionHandler) OnSubscribe(e SubscribeEvent) {
	if h.onSubscribe != nil {
		h.onSubscribe(e)
	}
}

func (h *testSubscriptionHandler) OnError(e SubscriptionErrorEvent) {
	if h.onError != nil {
		h.onError(e)
	}
}

func (h *testSubscriptionHandler) OnPublication(e PublicationEvent) {
	if h.onPublication != nil {
		h.onPublication(e)
	}
}

func (h *testSubscriptionHandler) OnUnsubscribe(e UnsubscribeEvent) {
	if h.onUnsubscribe != nil {
		h.onUnsubscribe(e)
	}
}

func TestConnectWrongAddress(t *testing.T) {
	client := NewJsonClient("ws://localhost:9000/connection/websocket", Config{})
	defer client.Close()
	doneCh := make(chan error, 1)
	handler := &testEventHandler{
		onError: func(e ErrorEvent) {
			var err TransportError
			if !errors.As(e.Error, &err) {
				doneCh <- fmt.Errorf("wrong error")
				return
			}
			close(doneCh)
		},
	}
	client.OnError(handler.OnError)
	_ = client.Connect()
	select {
	case err := <-doneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting disconnect due to malformed address")
	}
}

func TestSuccessfulConnect(t *testing.T) {
	client := NewProtobufClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	client.Close()
	doneCh := make(chan error, 1)
	handler := &testEventHandler{
		onConnect: func(e ConnectEvent) {
			if e.ClientID == "" {
				doneCh <- fmt.Errorf("wrong client ID value")
				return
			}
			close(doneCh)
		},
	}
	client.OnConnect(handler.OnConnect)
	_ = client.Connect()
	select {
	case err := <-doneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting successful connect")
	}
}

func TestDisconnect(t *testing.T) {
	client := NewProtobufClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	client.Close()
	connectDoneCh := make(chan error, 1)
	disconnectDoneCh := make(chan error, 1)
	handler := &testEventHandler{
		onConnect: func(e ConnectEvent) {
			close(connectDoneCh)
		},
		onDisconnect: func(e DisconnectEvent) {
			if e.Reconnect != false {
				disconnectDoneCh <- fmt.Errorf("wrong reconnect value")
				return
			}
			close(disconnectDoneCh)
		},
	}
	client.OnConnect(handler.OnConnect)
	client.OnDisconnect(handler.OnDisconnect)
	_ = client.Connect()
	select {
	case err := <-connectDoneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting successful connect")
	}
	client.Disconnect()
	select {
	case err := <-disconnectDoneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting successful disconnect")
	}
}

func TestPublishProtobuf(t *testing.T) {
	client := NewProtobufClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	client.Close()
	_ = client.Connect()
	_, err := client.Publish("test", []byte("boom"))
	if err != nil {
		t.Errorf("error publish: %v", err)
	}
}

func TestPublishJSON(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	_, err := client.Publish("test", []byte("{}"))
	if err != nil {
		t.Errorf("error publish: %v", err)
	}
}

func TestPublishInvalidJSON(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	_, err := client.Publish("test", []byte("boom"))
	if err == nil {
		t.Errorf("error expected on publish invalid JSON")
	}
}

func TestSubscribeSuccess(t *testing.T) {
	doneCh := make(chan error, 1)
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	sub, err := client.NewSubscription("test")
	if err != nil {
		t.Errorf("error on new subscription: %v", err)
	}
	subHandler := &testSubscriptionHandler{
		onSubscribe: func(e SubscribeEvent) {
			close(doneCh)
		},
	}
	sub.OnSubscribe(subHandler.OnSubscribe)
	_ = sub.Subscribe()
	select {
	case err := <-doneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting successful subscribe")
	}
}

func TestSubscribeError(t *testing.T) {
	doneCh := make(chan error, 1)
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	sub, err := client.NewSubscription("test:test")
	if err != nil {
		t.Errorf("error on new subscription: %v", err)
	}
	subHandler := &testSubscriptionHandler{
		onError: func(e SubscriptionErrorEvent) {
			// Due to unknown namespace.
			close(doneCh)
		},
	}
	sub.OnError(subHandler.OnError)
	_ = sub.Subscribe()
	select {
	case err := <-doneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting subscribe error")
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[random.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestHandlePublish(t *testing.T) {
	doneCh := make(chan error, 1)
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	sub, err := client.NewSubscription("test_handle_publish")
	if err != nil {
		t.Errorf("error on new subscription: %v", err)
	}
	msg := []byte(`{"unique":"` + randString(6) + strconv.FormatInt(time.Now().UnixNano(), 10) + `"}`)
	handler := &testSubscriptionHandler{
		onSubscribe: func(e SubscribeEvent) {
			_, err := client.Publish("test_handle_publish", msg)
			if err != nil {
				t.Fail()
			}
		},
		onPublication: func(e PublicationEvent) {
			if !bytes.Equal(e.Data, msg) {
				return
			}
			if e.Info == nil {
				doneCh <- fmt.Errorf("expecting non nil publication info")
				return
			}
			close(doneCh)
		},
	}
	sub.OnSubscribe(handler.OnSubscribe)
	sub.OnPublication(handler.OnPublication)
	_ = sub.Subscribe()
	select {
	case err := <-doneCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("expecting publication received over subscription")
	}
}

func TestSubscription_Unsubscribe(t *testing.T) {
	subscribedCh := make(chan struct{}, 1)
	unsubscribedCh := make(chan struct{}, 1)
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	sub, err := client.NewSubscription("test_subscription_close")
	if err != nil {
		t.Errorf("error on new subscription: %v", err)
	}
	handler := &testSubscriptionHandler{
		onSubscribe: func(e SubscribeEvent) {
			close(subscribedCh)
		},
		onUnsubscribe: func(event UnsubscribeEvent) {
			close(unsubscribedCh)
		},
	}
	sub.OnUnsubscribe(handler.OnUnsubscribe)
	sub.OnSubscribe(handler.OnSubscribe)
	sub.OnPublication(handler.OnPublication)
	_ = sub.Subscribe()
	select {
	case <-subscribedCh:
		if err != nil {
			t.Errorf("finish with error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("timeout waiting for subscribe")
	}
	err = sub.Unsubscribe()
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	select {
	case <-unsubscribedCh:
	case <-time.After(3 * time.Second):
		t.Errorf("timeout waiting for subscribe")
	}
}

func TestClient_Publish(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	msg := []byte(`{"unique":"` + randString(6) + strconv.FormatInt(time.Now().UnixNano(), 10) + `"}`)
	_, err := client.Publish("test", msg)
	if err != nil {
		// Publish should be allowed since we are using Centrifugo in insecure mode in tests.
		t.Fatal(err)
	}
}

func TestClient_Presence(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	_, err := client.Presence("test")
	var e *Error
	if !errors.As(err, &e) {
		t.Fatal("expected protocol error")
	}
	if e.Code != 108 {
		t.Fatal("expected not available error, got " + strconv.FormatUint(uint64(e.Code), 10))
	}
}

func TestClient_PresenceStats(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	_, err := client.PresenceStats("test")
	var e *Error
	if !errors.As(err, &e) {
		t.Fatal("expected protocol error")
	}
	if e.Code != 108 {
		t.Fatal("expected not available error, got " + strconv.FormatUint(uint64(e.Code), 10))
	}
}

func TestClient_History(t *testing.T) {
	client := NewJsonClient("ws://localhost:8000/connection/websocket?cf_protocol_version=v2", Config{})
	defer client.Close()
	_ = client.Connect()
	_, err := client.History("test")
	var e *Error
	if !errors.As(err, &e) {
		t.Fatal("expected protocol error")
	}
	if e.Code != 108 {
		t.Fatal("expected not available error, got " + strconv.FormatUint(uint64(e.Code), 10))
	}
}
