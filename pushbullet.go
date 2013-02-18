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
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

const HOST = "https://www.pushbullet.com/api"

// A Client connects to PushBullet with an API Key.
type Client struct {
	Key string
}

// New creates a new client with your personal API key.
func New(apikey string) *Client {
	return &Client{apikey}
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

func (c *Client) buildQuery() string {
	u, err := url.Parse(HOST)
	if err != nil {
		panic(err)
	}
	u.User = url.User(c.Key)
	return u.String()
}

// Devices fetches a list of devices from PushBullet.
func (c *Client) Devices() ([]*Device, error) {
	resp, err := http.Get(c.buildQuery() + "/devices")
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
	resp, err := http.PostForm(c.buildQuery()+"/pushes", data)
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
