go-pushbullet
=============

[![Build](https://img.shields.io/travis/xconstruct/go-pushbullet.svg?style=flat-square)](https://travis-ci.org/xconstruct/go-pushbullet)
[![API Documentation](https://img.shields.io/badge/api-GoDoc-blue.svg?style=flat-square)](https://godoc.org/github.com/xconstruct/go-pushbullet)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](http://opensource.org/licenses/MIT)

Simple Go client for [Pushbullet](http://pushbullet.com), a webservice to push
lists, addresses, links and more to your Android devices.

Documentation available under: http://godoc.org/github.com/xconstruct/go-pushbullet

### Install ###

	go get "github.com/xconstruct/go-pushbullet"

### Example ###

```go
pb := pushbullet.New("YOUR_API_KEY")
devs, err := pb.Devices()
if err != nil {
	panic(err)
}

err = pb.PushNote(devs[0].Iden, "Hello!", "Hi from go-pushbullet!")
if err != nil {
	panic(err)
}


user, err := pb.Me()
if err != nil {
	panic(err)
}

err = pb.PushSMS(user.Iden, devs[0].Iden, "<TARGET_PHONE_NUMBER>", "Sms text")
if err != nil {
	panic(err)
}

You can also retrieve a pushbullet device by nickname, and call the same methods as you would with the pushbbulet client
dev, err := pb.Device("My Phone")
if err != nil {
	panic(err)
}

err = dev.PushNote("Hello!", "Straight to device with just a title and body")
if err != nil {
	panic(err)
}

Channels are also supported in a similar manner
subs, err := pb.Subscriptions()
if err != nil {
	panic(err)
}

err = pb.PushNoteToChannel(subs[0].Channel.Tag, "Hello!", "Hi from go-pushbullet!")
if err != nil {
	panic(err)
}

sub, err := pb.Subscription("MyChannelTag")
if err != nil {
	panic(err)
}

err = sub.PushNote("Hello!", "Straight to Channel with just a title and body")
if err != nil {
	panic(err)
}

```
