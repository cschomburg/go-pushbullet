package pushbullet_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	pushbullet "github.com/xconstruct/go-pushbullet"
)

var TestDevice = &pushbullet.Device{
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

var TestError = &pushbullet.ErrResponse{
	Type:    "invalid_request",
	Message: "The resource could not be found.",
	Cat:     "~(=^‥^)",
}

func PushbulletResponseStub() *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/devices":
			d, _ := json.Marshal(TestDevice)
			resp = `{ "devices": [` + string(d) + `] }`
		default:
			e, _ := json.Marshal(TestError)
			resp := `{ "error":` + string(e) + `}`
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, resp, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
}

func PushbulletErrResponseStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(TestError)
		resp := `{ "error":` + string(e) + `}`
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, resp, http.StatusInternalServerError)
	}))
}

func TestDevices(t *testing.T) {
	server := PushbulletResponseStub()
	defer server.Close()
	pb := pushbullet.New("API_KEY")
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.NoError(t, err)
	assert.Len(t, devs, 1)
	assert.Equal(t, TestDevice.Nickname, devs[0].Nickname)
}

func TestDevicesError(t *testing.T) {
	server := PushbulletErrResponseStub()
	defer server.Close()
	pb := pushbullet.New("API_KEY")
	pb.Endpoint.URL = server.URL
	devs, err := pb.Devices()
	assert.Error(t, err)
	assert.Len(t, devs, 0)
	assert.Equal(t, TestError, err)
}
