package pushbullet

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var d = &Device{
	Active:       true,
	AppVersion:   8623,
	Created:      1.412047948579029e+09,
	Iden:         "ujpah72o0sjAoRtnM0jc",
	Manufacturer: "Apple",
	Model:        "iPhone 5s (GSM)",
	Modified:     1.412047948579031e+09,
	Nickname:     "Elon Musk's iPhone",
	PushToken:    "production:f73be0ee7877c8c7fa69b1468cde764f",
}

var n = &Note{
	Type:  "note",
	Title: "Space Travel Ideas",
	Body:  "Space Elevator, Mars Hyperloop, Space Model S (Model Space?)",
}

var e = &ErrResponse{
	Type:    "invalid_request",
	Message: "The resource could not be found.",
	Cat:     "~(=^‥^)",
}

var m = &User{
	Created:         1.381092887398433e+09,
	Email:           "elon@teslamotors.com",
	EmailNormalized: "elon@teslamotors.com",
	Iden:            "ujpah72o0",
	ImageUrl:        "https://static.pushbullet.com/missing-image/55a7dc-45",
	Modified:        1.441054560741007e+09,
	Name:            "Elon Musk",
}

var l = &Link{
	Type:  "link",
	Title: "Google",
	Body:  "Google homepage",
	URL:   "https://www.google.com",
}

var s = &EphemeralPush{
	Type:             "messaging_extension_reply",
	PackageName:      "com.pushbullet.android",
	SourceUserIden:   "ujpah72o0",
	TargetDeviceIden: "ujpah72o0sjAoRtnM0jc",
	ConversationIden: "+1 303 555 1212",
	Message:          "Hello!",
}

var k = "API_KEY"

func PushbulletResponseStub() *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/devices":
			d, _ := json.Marshal(d)
			resp = `{ "devices": [` + string(d) + `] }`
		case "/users/me":
			m, _ := json.Marshal(m)
			resp = string(m)
		case "/pushes":
			n, _ := json.Marshal(n)
			resp = string(n)
		case "/ephemerals":
			s, _ := json.Marshal(s)
			resp = string(s)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
}

func PushbulletErrJSONResponseStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(e)
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
	pb := New(k)
	assert.Equal(t, k, pb.Key)
}

func TestNewWithClient(t *testing.T) {
	c := &http.Client{}
	pb := NewWithClient(k, c)
	assert.Equal(t, c, pb.Client)
}

func TestError(t *testing.T) {
	assert.Equal(t, e.Message, e.Error())
}

func TestBuildRequest(t *testing.T) {
	pb := New(k)
	req := pb.buildRequest("/pushes", n)
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	var note Note
	json.Unmarshal(buf.Bytes(), &note)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, n, &note)
}

func TestDevices(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	d.Client = pb
	devs, err := pb.Devices()
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	assert.Equal(t, d, devs[0])
}

func TestDeviceWithNickname(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	dev, err := pb.Device(d.Nickname)
	assert.NoError(t, err)
	assert.Equal(t, d.Nickname, dev.Nickname)
	assert.Equal(t, pb, dev.Client)
}

func TestDeviceWithNicknameMissing(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	dev, err := pb.Device("MISSING")
	assert.Error(t, err)
	assert.Nil(t, dev)
}

func TestDevicePushNote(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(d.Nickname)
	err := dev.PushNote(n.Title, n.Body)
	assert.NoError(t, err)
}

func TestDevicePushLink(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(d.Nickname)
	err := dev.PushLink(l.Title, l.URL, l.Body)
	assert.NoError(t, err)
}

func TestDevicePushSMS(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	dev, _ := pb.Device(d.Nickname)
	err := dev.PushSMS(s.TargetDeviceIden, s.ConversationIden, s.Message)
	assert.NoError(t, err)
}

func TestDevicesError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.Error(t, err)
	assert.Len(t, devs, 0)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestDevicesJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.Error(t, err)
	assert.Len(t, devs, 0)
	assert.Equal(t, e, err)
}

func TestMe(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	me, err := pb.Me()
	assert.NoError(t, err)
	assert.Equal(t, m, me)
}

func TestMeError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	_, err := pb.Me()
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestMeJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	_, err := pb.Me()
	assert.Error(t, err)
	assert.Equal(t, e, err)
}

func TestPush(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", n)
	assert.NoError(t, err)
}

func TestPushError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", n)
	assert.Error(t, err)
	assert.Equal(t, "500 Internal Server Error", err.Error())
}

func TestPushJSONError(t *testing.T) {
	server := PushbulletErrJSONResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.Push("/pushes", n)
	assert.Error(t, err)
	assert.Equal(t, e, err)
}

func TestPushLink(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.PushLink(m.Iden, l.Title, l.URL, l.Body)
	assert.NoError(t, err)
}

func TestPushNote(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.PushNote(m.Iden, n.Title, n.Body)
	assert.NoError(t, err)
}

func TestPushSMS(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := New(k)
	pb.Endpoint.URL = server.URL
	err := pb.PushSMS(s.SourceUserIden, s.TargetDeviceIden, s.ConversationIden, s.Message)
	assert.NoError(t, err)
}
