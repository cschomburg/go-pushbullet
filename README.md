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
```
