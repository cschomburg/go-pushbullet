// Package pushbullet provides simple access to the v2 API of http://pushbullet.com.
/*

Example client:
	pb := pushbullet.New("YOUR_API_KEY")
	devices, err := pb.Devices()
	...
	err = pb.PushNote(devices[0].Iden, "Hello!", "Hi from go-pushbullet!")

The API is document at https://docs.pushbullet.com/http/ .
At the moment, it only supports querying devices and sending notifications.

*/
package pushbullet

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// AllDevices can be used an iden for sending a push to all devices
	AllDevices = ""
)

// ErrDeviceNotFound is raised when device nickname is not found on pusbullet server
var ErrDeviceNotFound = errors.New("Device not found")

// EndpointURL sets the default URL for the Pushbullet API
var EndpointURL = "https://api.pushbullet.com/v2"

// Endpoint allows manipulation of pushbullet API endpoint for testing
type Endpoint struct {
	URL string
}

// A Client connects to PushBullet with an API Key.
type Client struct {
	Key    string
	Client *http.Client
	Endpoint
}

// New creates a new client with your personal API key.
func New(apikey string) *Client {
	endpoint := Endpoint{URL: EndpointURL}
	return &Client{apikey, http.DefaultClient, endpoint}
}

// NewWithClient creates a new client with your personal API key and the given http Client
func NewWithClient(apikey string, client *http.Client) *Client {
	endpoint := Endpoint{URL: EndpointURL}
	return &Client{apikey, client, endpoint}
}

// A Device is a PushBullet device
type Device struct {
	Iden              string  `json:"iden"`
	Active            bool    `json:"active"`
	Created           float32 `json:"created"`
	Modified          float32 `json:"modified"`
	Icon              string  `json:"icon"`
	Nickname          string  `json:"nickname"`
	GeneratedNickname bool    `json:"generated_nickname"`
	Manufacturer      string  `json:"manufacturer"`
	Model             string  `json:"model"`
	AppVersion        int     `json:"app_version"`
	Fingerprint       string  `json:"fingerprint"`
	KeyFingerprint    string  `json:"key_fingerprint"`
	PushToken         string  `json:"push_token"`
	HasSms            bool    `json:"has_sms"`
	Client            *Client `json:"-"`
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

type subscriptionResponse struct {
	Subscriptions []*Subscription
}

func (client *Client) buildRequest(ctx context.Context, object string, data interface{}) *http.Request {
	r, err := http.NewRequestWithContext(ctx, "GET", client.Endpoint.URL+object, nil)
	if err != nil {
		panic(err)
	}

	// appengine sdk requires us to set the auth header by hand
	u := url.UserPassword(client.Key, "")
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(u.String())))

	if data != nil {
		r.Method = "POST"
		r.Header.Set("Content-Type", "application/json")
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		_ = enc.Encode(data)
		r.Body = ioutil.NopCloser(&b)
	}

	return r
}

// Devices fetches a list of devices from PushBullet.
func (client *Client) Devices() ([]*Device, error) {
	return client.DevicesWithContext(context.Background())
}

// DevicesWithContext fetches a list of devices from PushBullet.
func (client *Client) DevicesWithContext(ctx context.Context) (devices []*Device, retErr error) {
	req := client.buildRequest(ctx, "/devices", nil)
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			devices = nil
			retErr = fmt.Errorf("Unable to close connection to PushBullet: %w", closeErr)
		}
	}()

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

	devices = devResp.Devices
	for i := range devices {
		devices[i].Client = client
	}

	devices = append(devices, devResp.SharedDevices...)

	return devices, nil
}

// Device fetches an device with a given nickname from PushBullet.
func (client *Client) Device(nickname string) (*Device, error) {
	return client.DeviceWithContext(context.Background(), nickname)
}

// DeviceWithContext fetches an device with a given nickname from PushBullet.
func (client *Client) DeviceWithContext(ctx context.Context, nickname string) (*Device, error) {
	devices, err := client.DevicesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].Nickname == nickname {
			devices[i].Client = client
			return devices[i], nil
		}
	}
	return nil, ErrDeviceNotFound
}

// PushNote sends a note to the specific device with the given title and body
func (device *Device) PushNote(title string, body string) error {
	return device.PushNoteWithContext(context.Background(), title, body)
}

// PushNoteWithContext sends a note to the specific device with the given title and body
func (device *Device) PushNoteWithContext(ctx context.Context, title string, body string) error {
	return device.Client.PushNote(device.Iden, title, body)
}

// PushLink sends a link to the specific device with the given title and url
func (device *Device) PushLink(title string, u string, body string) error {
	return device.PushLinkWithContext(context.Background(), title, u, body)
}

// PushLinkWithContext sends a link to the specific device with the given title and url
func (device *Device) PushLinkWithContext(ctx context.Context, title string, u string, body string) error {
	return device.Client.PushLink(device.Iden, title, u, body)
}

// PushSMS sends an SMS to the specific user from the device with the given title and url
func (device *Device) PushSMS(deviceIden string, phoneNumber string, message string) error {
	return device.PushSMSWithContext(context.Background(), deviceIden, phoneNumber, message)
}

// PushSMSWithContext sends an SMS to the specific user from the device with the given title and url
func (device *Device) PushSMSWithContext(
	ctx context.Context,
	deviceIden string,
	phoneNumber string,
	message string,
) error {
	return device.Client.PushSMS(device.Iden, deviceIden, phoneNumber, message)
}

// User represents the User object for pushbullet
type User struct {
	Iden            string      `json:"iden"`
	Email           string      `json:"email"`
	EmailNormalized string      `json:"email_normalized"`
	Created         float64     `json:"created"`
	Modified        float64     `json:"modified"`
	Name            string      `json:"name"`
	ImageURL        string      `json:"image_url"`
	Preferences     interface{} `json:"preferences"`
}

// Me returns the user object for the pushbullet user
func (client *Client) Me() (*User, error) {
	return client.MeWithContext(context.Background())
}

// MeWithContext returns the user object for the pushbullet user
func (client *Client) MeWithContext(ctx context.Context) (user *User, retErr error) {
	req := client.buildRequest(ctx, "/users/me", nil)
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			retErr = fmt.Errorf("Unable to close connection to PushBullet: %w", closeErr)
		}
	}()

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
func (client *Client) Push(endPoint string, data interface{}) (retErr error) {
	return client.PushWithContext(context.Background(), endPoint, data)
}

// PushWithContext pushes the data to a specific device registered with PushBullet.  The
// 'data' parameter is marshaled to JSON and sent as the request body.  Most
// users should call one of PusNote, PushLink, PushAddress, or PushList.
func (client *Client) PushWithContext(
	ctx context.Context,
	endPoint string,
	data interface{},
) (retErr error) {
	req := client.buildRequest(ctx, endPoint, data)
	resp, err := client.Client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			retErr = fmt.Errorf("Unable to close connection to PushBullet: %w", closeErr)
		}
	}()

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

// Note exposes the required and optional fields of the Pushbullet push type=note
type Note struct {
	Iden  string `json:"device_iden,omitempty"`
	Tag   string `json:"channel_tag,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// PushNote pushes a note with title and body to a specific PushBullet device.
func (client *Client) PushNote(iden string, title string, body string) error {
	return client.PushNoteWithContext(context.Background(), iden, title, body)
}

// PushNoteWithContext pushes a note with title and body to a specific PushBullet device.
func (client *Client) PushNoteWithContext(
	ctx context.Context,
	iden string,
	title string,
	body string,
) error {
	data := Note{
		Iden:  iden,
		Type:  "note",
		Title: title,
		Body:  body,
	}

	return client.PushWithContext(ctx, "/pushes", data)
}

// PushNoteToChannel pushes a note with title and body to a specific PushBullet channel.
func (client *Client) PushNoteToChannel(tag string, title string, body string) error {
	return client.PushNoteToChannelWithContext(context.Background(), tag, title, body)
}

// PushNoteToChannelWithContext pushes a note with title and body to a specific PushBullet channel.
func (client *Client) PushNoteToChannelWithContext(
	ctx context.Context,
	tag string,
	title string,
	body string,
) error {
	data := Note{
		Tag:   tag,
		Type:  "note",
		Title: title,
		Body:  body,
	}

	return client.PushWithContext(ctx, "/pushes", data)
}

// Link exposes the required and optional fields of the Pushbullet push type=link
type Link struct {
	Iden  string `json:"device_iden,omitempty"`
	Tag   string `json:"channel_tag,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body,omitempty"`
}

// PushLink pushes a link with a title and url to a specific PushBullet device.
func (client *Client) PushLink(iden string, title string, u string, body string) error {
	return client.PushLinkWithContext(context.Background(), iden, title, u, body)
}

// PushLinkWithContext pushes a link with a title and url to a specific PushBullet device.
func (client *Client) PushLinkWithContext(
	ctx context.Context,
	iden string,
	title string,
	u string,
	body string,
) error {
	data := Link{
		Iden:  iden,
		Type:  "link",
		Title: title,
		URL:   u,
		Body:  body,
	}

	return client.PushWithContext(ctx, "/pushes", data)
}

// PushLinkToChannel pushes a link with a title and url to a specific PushBullet device.
func (client *Client) PushLinkToChannel(tag string, title string, u string, body string) error {
	return client.PushLinkToChannelWithContext(context.Background(), tag, title, u, body)
}

// PushLinkToChannelWithContext pushes a link with a title and url to a specific PushBullet device.
func (client *Client) PushLinkToChannelWithContext(
	ctx context.Context,
	tag string,
	title string,
	u string,
	body string,
) error {
	data := Link{
		Tag:   tag,
		Type:  "link",
		Title: title,
		URL:   u,
		Body:  body,
	}

	return client.PushWithContext(ctx, "/pushes", data)
}

// EphemeralPush  exposes the required fields of the Pushbullet ephemeral object
type EphemeralPush struct {
	Type             string `json:"type"`
	PackageName      string `json:"package_name"`
	SourceUserIden   string `json:"source_user_iden"`
	TargetDeviceIden string `json:"target_device_iden"`
	ConversationIden string `json:"conversation_iden"`
	Message          string `json:"message"`
}

// Ephemeral constructs the Ephemeral object for pushing which requires the EphemeralPush object
type Ephemeral struct {
	Type string        `json:"type"`
	Push EphemeralPush `json:"push"`
}

// PushSMS sends an SMS message with pushbullet
func (client *Client) PushSMS(
	userIden string,
	deviceIden string,
	phoneNumber string,
	message string,
) error {
	return client.PushSMSWithContext(
		context.Background(),
		userIden,
		deviceIden,
		phoneNumber,
		message,
	)
}

// PushSMSWithContext sends an SMS message with pushbullet
func (client *Client) PushSMSWithContext(
	ctx context.Context,
	userIden string,
	deviceIden string,
	phoneNumber string,
	message string,
) error {
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

	return client.PushWithContext(ctx, "/ephemerals", data)
}

// Subscription object allows interaction with pushbullet channels
type Subscription struct {
	Iden     string   `json:"iden"`
	Active   bool     `json:"active"`
	Created  float32  `json:"created"`
	Modified float32  `json:"modified"`
	Muted    string   `json:"muted"`
	Channel  *Channel `json:"channel"`
	Client   *Client  `json:"-"`
}

// Channel object contains specific information about the pushbullet Channel
type Channel struct {
	Iden        string `json:"iden"`
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	WebsiteURL  string `json:"website_url"`
}

// Subscriptions gets the list of subscriptions.
func (client *Client) Subscriptions() (subscriptions []*Subscription, retErr error) {
	return client.SubscriptionsWithContext(context.Background())
}

// SubscriptionsWithContext gets the list of subscriptions.
func (client *Client) SubscriptionsWithContext(ctx context.Context) (subscriptions []*Subscription, retErr error) {
	req := client.buildRequest(ctx, "/subscriptions", nil)
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			retErr = fmt.Errorf("Unable to close connection to PushBullet: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errjson errorResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&errjson)
		if err == nil {
			return nil, &errjson.ErrResponse
		}

		return nil, errors.New(resp.Status)
	}

	var subResp subscriptionResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&subResp)
	if err != nil {
		return nil, err
	}

	for i := range subResp.Subscriptions {
		subResp.Subscriptions[i].Client = client
	}
	subscriptions = subResp.Subscriptions
	return subscriptions, nil
}

// Subscription fetches an subscription with a given channel tag from PushBullet.
func (client *Client) Subscription(tag string) (*Subscription, error) {
	return client.SubscriptionWithContext(context.Background(), tag)
}

// SubscriptionWithContext fetches an subscription with a given channel tag from PushBullet.
func (client *Client) SubscriptionWithContext(ctx context.Context, tag string) (*Subscription, error) {
	subs, err := client.SubscriptionsWithContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range subs {
		if subs[i].Channel.Tag == tag {
			subs[i].Client = client
			return subs[i], nil
		}
	}
	return nil, ErrDeviceNotFound
}

// PushNote sends a note to the specific Channel with the given title and body
func (subscription *Subscription) PushNote(title string, body string) error {
	return subscription.PushNoteWithContext(context.Background(), title, body)
}

// PushNoteWithContext sends a note to the specific Channel with the given title and body
func (subscription *Subscription) PushNoteWithContext(ctx context.Context, title string, body string) error {
	return subscription.Client.PushNoteToChannel(subscription.Channel.Tag, title, body)
}

// PushLink sends a link to the specific Channel with the given title, url and body
func (subscription *Subscription) PushLink(title string, u string, body string) error {
	return subscription.PushLinkWithContext(context.Background(), title, u, body)
}

// PushLinkWithContext sends a link to the specific Channel with the given title, url and body
func (subscription *Subscription) PushLinkWithContext(
	ctx context.Context,
	title string,
	u string,
	body string,
) error {
	return subscription.Client.PushLinkToChannel(subscription.Channel.Tag, title, u, body)
}
