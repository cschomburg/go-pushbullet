/*
Package pushbullet provides simple access to the API of http://pushbullet.com.

Example client:
	pb := pushbullet.New("YOUR_API_KEY")
	devices, err := pb.Devices()
	...
	err = pb.PushNote(devices[0].Id, "Hello!", "Hi from go-pushbullet!")
*/
package pushbullet

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const HOST = "https://www.pushbullet.com/api"

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

// A Device represents an Android Device as reported by PushBullet.
type Device struct {
	Id        int
	OwnerName string `json:"owner_name"`
	Extras    struct {
		Manufacturer   string
		Model          string
		AndroidVersion string `json:"android_version"`
		SdkVersion     string `json:"sdk_version"`
		AppVersion     string `json:"app_version"`
		Nickname       string
	}
}

type deviceResponse struct {
	Devices       []*Device
	SharedDevices []*Device `json:"shared_devices"`
}

func (c *Client) buildRequest(endpoint string, data url.Values) *http.Request {
	r, err := http.NewRequest("GET", HOST+endpoint, nil)
	if err != nil {
		panic(err)
	}

	// appengine sdk requires us to set the auth header by hand
	u := url.UserPassword(c.Key, "")
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(u.String())))

	if data != nil {
		r.Method = "POST"
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Body = ioutil.NopCloser(strings.NewReader(data.Encode()))
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
	defer func() {
		resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
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

// Push pushes the data to a specific device registered with PushBullet.
func (c *Client) Push(deviceId int, data url.Values) error {
	data.Set("device_id", strconv.Itoa(deviceId))
	req := c.buildRequest("/pushes", data)
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

// PushNote pushes a note with title and body to a specific PushBullet device.
func (c *Client) PushNote(deviceId int, title, body string) error {
	data := url.Values{
		"type":  {"note"},
		"title": {title},
		"body":  {body},
	}
	return c.Push(deviceId, data)
}

// PushAddress pushes a geo address with name and address to a specific PushBullet device.
func (c *Client) PushAddress(deviceId int, name, address string) error {
	data := url.Values{
		"type":    {"address"},
		"name":    {name},
		"address": {address},
	}
	return c.Push(deviceId, data)
}

// PushList pushes a list with name and a slice of items to a specific PushBullet device.
func (c *Client) PushList(deviceId int, title string, items []string) error {
	data := url.Values{
		"type":  {"list"},
		"title": {title},
		"items": items,
	}
	return c.Push(deviceId, data)
}

// PushLink pushes a link with a title and url to a specific PushBullet device.
func (c *Client) PushLink(deviceId int, title, u string) error {
	data := url.Values{
		"type":  {"link"},
		"title": {title},
		"url":   {u},
	}
	return c.Push(deviceId, data)
}
