package main

import (
	"github.com/xconstruct/go-pushbullet"
)

func main() {
	pb := pushbullet.New("YOUR_API_KEY")
	devs, err := pb.Devices()
	if err != nil {
		panic(err)
	}

	err = pb.PushNote(devs[0].Iden, "Hello!", "Hi from go-pushbullet!")
	if err != nil {
		panic(err)
	}
}
