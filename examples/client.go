package main

import (
	"github.com/durandj/go-pushbullet"
)

func main() {
	pb := pushbullet.New("<YOUR_API_KEY>")
	devs, err := pb.Devices()
	if err != nil {
		panic(err)
	}

	err = pb.PushNote(devs[0].Iden, "Hello!", "Hi from go-pushbullet!")
	if err != nil {
		panic(err)
	}

	//SMS test
	user, err := pb.Me()
	if err != nil {
		panic(err)
	}

	err = pb.PushSMS(user.Iden, devs[0].Iden, "<TARGET_PHONE_NUMBER>", "Sms text")
	if err != nil {
		panic(err)
	}
}
