// Package pushbullet provides simple access to the v2 API of http://pushbullet.com.
/*

Example client:
	pb := pushbullet.New("YOUR_API_KEY")
	devices, err := pb.Devices()
	...
	err = pb.PushNote(devices[0].Iden, "Hello!", "Hi from go-pushbullet!")

The API is document at https://docs.pushbullet.com/http/ .  At the moment, it only supports querying devices and sending notifications.

*/
package pushbullet

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var Endpoint = "https://api.pushbullet.com/v2"

// A Client connects to PushBullet with an API Key.
type Client struct {
	Key    string
	Client *http.Client
}

// New creates a new client with your personal API key.
func New(apikey string) *Client {
	return &Client{apikey, http.DefaultClient}
}

// New creates a new client with your personal API key and the given http Client
func NewWithClient(apikey string, client *http.Client) *Client {
	return &Client{apikey, client}
}

// A Device is a PushBullet device
type Device struct {
	Iden         string  `json:"iden"`
	PushToken    string  `json:"push_token"`
	AppVersion   int     `json:"app_version"`
	Fingerprint  string  `json:"fingerprint"`
	Active       bool    `json:"active"`
	Nickname     string  `json:"nickname"`
	Manufacturer string  `json:"manufacturer"`
	Type         string  `json:"type"`
	Created      float32 `json:"created"`
	Modified     float32 `json:"modified"`
	Model        string  `json:"model"`
	Pushable     bool    `json:"pushable"`
}

// ErrResponse is an error returned by the PushBullet API
type ErrResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Cat     string `json:"cat"`
}

func (e *ErrResponse) Error() string {
	return e.Message
}

type errorResponse struct {
	ErrResponse `json:"error"`
}

type deviceResponse struct {
	Devices       []*Device
	SharedDevices []*Device `json:"shared_devices"`
}

func (c *Client) buildRequest(object string, data interface{}) *http.Request {
	r, err := http.NewRequest("GET", Endpoint+object, nil)
	if err != nil {
		panic(err)
	}

	// appengine sdk requires us to set the auth header by hand
	u := url.UserPassword(c.Key, "")
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(u.String())))

	if data != nil {
		r.Method = "POST"
		r.Header.Set("Content-Type", "application/json")
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		enc.Encode(data)
		r.Body = ioutil.NopCloser(&b)
	}

	return r
}

// Devices fetches a list of devices from PushBullet.
func (c *Client) Devices() ([]*Device, error) {
	req := c.buildRequest("/devices", nil)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errjson errorResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&errjson)
		if err == nil {
			return nil, &errjson.ErrResponse
		}

		return nil, errors.New(resp.Status)
	}

	var devResp deviceResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&devResp)
	if err != nil {
		return nil, err
	}

	devices := append(devResp.Devices, devResp.SharedDevices...)
	return devices, nil
}

type User struct {
	Iden            string      `json:"iden"`
	Email           string      `json:"email"`
	EmailNormalized string      `json:"email_normalized"`
	Created         float64     `json:"created"`
	Modified        float64     `json:"modified"`
	Name            string      `json:"name"`
	ImageUrl        string      `json:"image_url"`
	Preferences     interface{} `json:"preferences"`
}

func (c *Client) Me() (*User, error) {
	req := c.buildRequest("/users/me", nil)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errjson errorResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&errjson)
		if err == nil {
			return nil, &errjson.ErrResponse
		}

		return nil, errors.New(resp.Status)
	}

	var userResponse User
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&userResponse)
	if err != nil {
		return nil, err
	}
	return &userResponse, nil
}

// Push pushes the data to a specific device registered with PushBullet.  The
// 'data' parameter is marshaled to JSON and sent as the request body.  Most
// users should call one of PusNote, PushLink, PushAddress, or PushList.
func (c *Client) Push(endPoint string, data interface{}) error {
	req := c.buildRequest(endPoint, data)
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResponse errorResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&errResponse)
		if err == nil {
			return &errResponse.ErrResponse
		}

		return errors.New(resp.Status)
	}

	return nil
}

type Note struct {
	Iden  string `json:"device_iden,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// PushNote pushes a note with title and body to a specific PushBullet device.
func (c *Client) PushNote(iden string, title, body string) error {
	data := Note{
		Iden:  iden,
		Type:  "note",
		Title: title,
		Body:  body,
	}
	return c.Push("/pushes", data)
}

type Address struct {
	Iden    string `json:"device_iden"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// PushAddress pushes a geo address with name and address to a specific PushBullet device.
func (c *Client) PushAddress(iden string, name, address string) error {
	data := Address{
		Iden:    iden,
		Type:    "address",
		Name:    name,
		Address: address,
	}
	return c.Push("/pushes", data)
}

type List struct {
	Iden  string   `json:"device_iden"`
	Type  string   `json:"type"`
	Title string   `json:"title"`
	Items []string `json:"items"`
}

// PushList pushes a list with name and a slice of items to a specific PushBullet device.
func (c *Client) PushList(iden string, title string, items []string) error {
	data := List{
		Iden:  iden,
		Type:  "list",
		Title: title,
		Items: items,
	}
	return c.Push("/pushes", data)
}

type Link struct {
	Iden  string `json:"device_iden"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body,omitempty"`
}

// PushLink pushes a link with a title and url to a specific PushBullet device.
func (c *Client) PushLink(iden, title, u, body string) error {
	data := Link{
		Iden:  iden,
		Type:  "link",
		Title: title,
		URL:   u,
		Body:  body,
	}
	return c.Push("/pushes", data)
}

type EphemeralPush struct {
	Type             string `json:"type"`
	PackageName      string `json:"package_name"`
	SourceUserIden   string `json:"source_user_iden"`
	TargetDeviceIden string `json:"target_device_iden"`
	ConversationIden string `json:"conversation_iden"`
	Message          string `json:"message"`
}

type Ephemeral struct {
	Type string        `json:"type"`
	Push EphemeralPush `json:"push"`
}

func (c *Client) PushSMS(userIden, deviceIden, phoneNumber, message string) error {
	data := Ephemeral{
		Type: "push",
		Push: EphemeralPush{
			Type:             "messaging_extension_reply",
			PackageName:      "com.pushbullet.android",
			SourceUserIden:   userIden,
			TargetDeviceIden: deviceIden,
			ConversationIden: phoneNumber,
			Message:          message,
		},
	}
	return c.Push("/ephemerals", data)
}
