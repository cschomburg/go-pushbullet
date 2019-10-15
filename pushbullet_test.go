package pushbullet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockDevice = &Device{
	Active:            true,
	AppVersion:        8623,
	Created:           1.412047948579029e+09,
	Iden:              "ujpah72o0sjAoRtnM0jc",
	Manufacturer:      "Apple",
	Model:             "iPhone 5s (GSM)",
	Modified:          1.412047948579031e+09,
	Nickname:          "Elon Musk's iPhone",
	GeneratedNickname: true,
	PushToken:         "production:f73be0ee7877c8c7fa69b1468cde764f",
}

var mockNote = &Note{
	Type:  "note",
	Title: "Space Travel Ideas",
	Body:  "Space Elevator, Mars Hyperloop, Space Model S (Model Space?)",
}

var mockError = &ErrResponse{
	Type:    "invalid_request",
	Message: "The resource could not be found.",
	Cat:     "~(=^‥^)",
}

var mockUser = &User{
	Created:         1.381092887398433e+09,
	Email:           "elon@teslamotors.com",
	EmailNormalized: "elon@teslamotors.com",
	Iden:            "ujpah72o0",
	ImageURL:        "https://static.pushbullet.com/missing-image/55a7dc-45",
	Modified:        1.441054560741007e+09,
	Name:            "Elon Musk",
}

var mockLink = &Link{
	Type:  "link",
	Title: "Google",
	Body:  "Google homepage",
	URL:   "https://www.google.com",
}

var mockSMS = &EphemeralPush{
	Type:             "messaging_extension_reply",
	PackageName:      "com.pushbullet.android",
	SourceUserIden:   "ujpah72o0",
	TargetDeviceIden: "ujpah72o0sjAoRtnM0jc",
	ConversationIden: "+1 303 555 1212",
	Message:          "Hello!",
}

var mockChannel = &Channel{
	Iden:        "ujxPklLhvyKsjAvkMyTVh6",
	Tag:         "elonmusknews",
	Name:        "Elon Musk News",
	Description: "News about Elon Musk.",
	ImageURL:    "https://dl.pushbulletusercontent.com/StzRmwdkIe8gluBH3XoJ9HjRqjlUYSf4/musk.jpg",
	WebsiteURL:  "http://elonmuscknews.com",
}

var mockSubscription = &Subscription{
	Active:   true,
	Channel:  mockChannel,
	Created:  1.412047948579029e+09,
	Iden:     "ujpah72o0sjAoRtnM0jc",
	Modified: 1.412047948579031e+09,
}

var apiKey = "API_KEY"

func PushbulletResponseStub() *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/devices":
			d, _ := json.Marshal(mockDevice)
			resp = `{ "devices": [` + string(d) + `] }`
		case "/users/me":
			m, _ := json.Marshal(mockUser)
			resp = string(m)
		case "/pushes":
			n, _ := json.Marshal(mockNote)
			resp = string(n)
		case "/ephemerals":
			s, _ := json.Marshal(mockSMS)
			resp = string(s)
		case "/subscriptions":
			sub, _ := json.Marshal(mockSubscription)
			resp = `{ "subscriptions": [` + string(sub) + `] }`
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
}

func PushbulletErrJSONResponseStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(mockError)
		resp := `{ "error":` + string(e) + `}`
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, resp, http.StatusInternalServerError)
	}))
}

func PushbulletErrResponseStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
}

func TestNew(t *testing.T) {
	pb := New(apiKey)
	assert.Equal(t, apiKey, pb.Key)
}

func TestNewWithClient(t *testing.T) {
	c := &http.Client{}
	pb := NewWithClient(apiKey, c)
	assert.Equal(t, c, pb.Client)
}

func TestError(t *testing.T) {
	assert.Equal(t, mockError.Message, mockError.Error())
}

func TestBuildRequest(t *testing.T) {
	pb := New(apiKey)
	req := pb.buildRequest(context.Background(), "/pushes", mockNote)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(req.Body)
	var note Note
	_ = json.Unmarshal(buf.Bytes(), &note)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, mockNote, &note)
}

func TestDevices(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	mockDevice.Client = pb
	devs, err := pb.Devices()
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	assert.Equal(t, mockDevice, devs[0])
}

func TestDeviceWithNickname(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	dev, err := pb.Device(mockDevice.Nickname)
	assert.NoError(t, err)
	assert.Equal(t, mockDevice.Nickname, dev.Nickname)
	assert.Equal(t, pb, dev.Client)
}

func TestDeviceWithNicknameMissing(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	dev, err := pb.Device("MISSING")
	assert.Error(t, err)
	assert.Nil(t, dev)
}

func TestDevicePushNote(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(mockDevice.Nickname)
	err := dev.PushNote(mockNote.Title, mockNote.Body)
	assert.NoError(t, err)
}

func TestDevicePushLink(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(mockDevice.Nickname)
	err := dev.PushLink(mockLink.Title, mockLink.URL, mockLink.Body)
	assert.NoError(t, err)
}

func TestDevicePushSMS(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(mockDevice.Nickname)
	err := dev.PushSMS(mockSMS.TargetDeviceIden, mockSMS.ConversationIden, mockSMS.Message)
	assert.NoError(t, err)
}

func TestDevicesError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.Error(t, err)
	assert.Len(t, devs, 0)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestDevicesJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.Error(t, err)
	assert.Len(t, devs, 0)
	assert.Equal(t, mockError, err)
}

func TestMe(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	me, err := pb.Me()
	assert.NoError(t, err)
	assert.Equal(t, mockUser, me)
}

func TestMeError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	_, err := pb.Me()
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestMeJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	_, err := pb.Me()
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
}

func TestPush(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", mockNote)
	assert.NoError(t, err)
}

func TestPushError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", mockNote)
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestPushJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", mockNote)
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
}

func TestPushLink(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.PushLink(mockUser.Iden, mockLink.Title, mockLink.URL, mockLink.Body)
	assert.NoError(t, err)
}

func TestPushNote(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.PushNote(mockUser.Iden, mockNote.Title, mockNote.Body)
	assert.NoError(t, err)
}

func TestPushLinkToChannel(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.PushLinkToChannel(mockSubscription.Channel.Tag, mockLink.Title, mockLink.URL, mockLink.Body)
	assert.NoError(t, err)
}

func TestPushNoteToChannel(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.PushNoteToChannel(mockSubscription.Channel.Tag, mockNote.Title, mockNote.Body)
	assert.NoError(t, err)
}

func TestPushSMS(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	err := pb.PushSMS(mockSMS.SourceUserIden, mockSMS.TargetDeviceIden, mockSMS.ConversationIden, mockSMS.Message)
	assert.NoError(t, err)
}

func TestSubscriptions(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	mockSubscription.Client = pb
	subs, err := pb.Subscriptions()
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, mockSubscription, subs[0])
}

func TestSubscriptionsError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	subs, err := pb.Subscriptions()
	assert.Error(t, err)
	assert.Len(t, subs, 0)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestSubscriptionsJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	subs, err := pb.Subscriptions()
	assert.Error(t, err)
	assert.Len(t, subs, 0)
	assert.Equal(t, mockError, err)
}

func TestSubscriptionWithName(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	sub, err := pb.Subscription(mockChannel.Tag)
	assert.NoError(t, err)
	assert.Equal(t, mockChannel.Tag, sub.Channel.Tag)
	assert.Equal(t, pb, sub.Client)
}

func TestSubscriptionWithNameMissing(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	sub, err := pb.Subscription("MISSING")
	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestSubscriptionPushNote(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	sub, _ := pb.Subscription(mockSubscription.Channel.Tag)
	err := sub.PushNote(mockNote.Title, mockNote.Body)
	assert.NoError(t, err)
}

func TestSubscriptionPushLink(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(apiKey)
	pb.Endpoint.URL = server.URL
	sub, _ := pb.Subscription(mockSubscription.Channel.Tag)
	err := sub.PushLink(mockLink.Title, mockLink.URL, mockLink.Body)
	assert.NoError(t, err)
}
