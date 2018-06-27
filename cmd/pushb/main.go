package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xconstruct/go-pushbullet"
)

type Devices []string

var (
	devices Devices
)

func (ds *Devices) String() string {
	return strings.Join([]string(*ds), ",")
}

func (ds *Devices) Set(value string) error {
	*ds = append(*ds, value)
	return nil
}

type Config struct {
	ApiKey  string   `json:"api_key"`
	Devices []Device `json:"devices"`
}

type Device struct {
	Iden string `json:"iden"`
	Name string `json:"name"`
}

func getArg(i int, fallback string) string {
	if len(os.Args) <= i {
		return ""
	}
	return os.Args[i]
}

func main() {
	flag.Var(&devices, "d", "Specify target devices")
	flag.Parse()

	cmd := getArg(1, "")

	switch cmd {
	case "login":
		login()
	case "note":
		pushNote()
	case "link":
		pushLink()
	case "devices":
		listDevices()
	default:
		printHelp()
	}
}

func home() string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" && home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func login() {
	key := getArg(2, "")
	var cfg Config

	cfg.ApiKey = key
	cfg.Devices = make([]Device, 0)

	if key == "" {
		writeConfig(cfg)
		return
	}

	pb := pushbullet.New(key)
	devs, err := pb.Devices()
	if err != nil {
		log.Fatalln(err)
	}
	for _, dev := range devs {
		name := dev.Nickname
		if name == "" {
			name = dev.Model
		}
		cfg.Devices = append(cfg.Devices, Device{
			Iden: dev.Iden,
			Name: name,
		})
	}
	writeConfig(cfg)
}

func readConfig() (Config, error) {
	cfgfile := filepath.Join(home(), ".pushb.config.json")
	f, err := os.Open(cfgfile)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	if err = dec.Decode(&cfg); err != nil {
		return Config{}, err
	}
	if len(devices) > 0 {
		var ds []Device
		for _, v1 := range cfg.Devices {
			for _, v2 := range devices {
				if v1.Iden == v2 {
					ds = append(ds, v1)
				}
			}
		}
		cfg.Devices = ds
	}
	return cfg, nil
}

func writeConfig(cfg Config) {
	cfgfile := filepath.Join(home(), ".pushb.config.json")
	f, err := os.OpenFile(cfgfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err = enc.Encode(cfg); err != nil {
		log.Fatalln(err)
	}
}

func pushNote() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	title := getArg(2, "")
	body := getArg(3, "")
	if body == "" {
		body = title
		title = ""
	}

	if body == "-" {
		// read stdin
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalln(err)
		}
		body = string(b)
	}

	pb := pushbullet.New(cfg.ApiKey)
	err = pb.PushNote(cfg.Devices[0].Iden, title, body)
	if err != nil {
		log.Fatalln(err)
	}
}

func pushLink() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	title := getArg(2, "")
	link := getArg(3, "")
	if link == "" {
		link = title
		title = ""
	}
	pb := pushbullet.New(cfg.ApiKey)
	err = pb.PushLink(cfg.Devices[0].Iden, title, link, "")
	if err != nil {
		log.Fatalln(err)
	}
}

func listDevices() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	for _, d := range cfg.Devices {
		fmt.Printf("%10s\t%s\n", d.Iden, d.Name)
	}
}

func printHelp() {
	topic := getArg(2, "")

	switch topic {
	default:
		fmt.Println(`Pushb is a simple client for PushBullet.

Usage:
    pushb command [flags] [arguments]

Commands:
    login      Saves the api key in the config
    devices    Shows a list of registered devices
    help       Shows this help

    address    Pushes an address to a device
    link       Pushes a link to a device
    list       Pushes a list to a device
    note       Pushes a note to a device
	
Use "pushb help [topic] for more information about that topic.`)
	}
}
