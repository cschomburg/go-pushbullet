go-pushbullet
=============

Simple Go client for [PushBullet](http://pushbullet.com), a webservice to push
lists, addresses, links and more to your Android devices.

### Install ###

	go get "github.com/xconstruct/go-pushbullet"

### Example ###

```go
pb := pushbullet.New("YOUR_API_KEY")
devs, err := pb.Devices()
if err != nil {
	panic(err)
}

err = pb.PushNote(devs[0].Id, "Hello!", "Hi from go-pushbullet!")
if err != nil {
	panic(err)
}
```
