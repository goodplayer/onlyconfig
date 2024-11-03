package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/meidoworks/nekoq-component/configure/configapi"

	"github.com/goodplayer/onlyconfig/client"
)

var sel string
var optSel string
var group string
var key string
var outputFile string

var configFile string

type serverList []string

func (s *serverList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *serverList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var srvList serverList

func init() {
	flag.StringVar(&sel, "sel", "", "selectors string, single configuration")
	flag.StringVar(&optSel, "optsel", "", "optional selectors string, single configuration")
	flag.StringVar(&group, "group", "", "group, single configuration")
	flag.StringVar(&key, "key", "", "key, single configuration")
	flag.StringVar(&outputFile, "output", "", "output file path, single configuration")
	flag.StringVar(&configFile, "config", "", "Config file path. This will override other flags for single configuration.")
	flag.Var(&srvList, "server", "server list: -server http://srv1 -server http://srv2")
	flag.Parse()
}

type Config struct {
	ConfigList []struct {
		SelectorsString         string `toml:"selectors"`
		OptionalSelectorsString string `toml:"optional_selectors"`
		Group                   string `toml:"group"`
		Key                     string `toml:"key"`
		Output                  string `toml:"output"`
	} `toml:"config_list"`
}

func (c *Config) Validate() error {
	//FIXME add validation
	return nil
}

func main() {
	var config = new(Config)
	if configFile != "" {
		f := func() []byte {
			file, err := os.Open(configFile)
			if err != nil {
				log.Fatal(err)
			}
			defer func(file *os.File) {
				_ = file.Close()
			}(file)
			data, err := io.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}
			return data
		}
		content := f()
		if err := toml.Unmarshal(content, config); err != nil {
			log.Fatal(err)
		}
	} else {
		config.ConfigList = append(config.ConfigList, struct {
			SelectorsString         string `toml:"selectors"`
			OptionalSelectorsString string `toml:"optional_selectors"`
			Group                   string `toml:"group"`
			Key                     string `toml:"key"`
			Output                  string `toml:"output"`
		}{SelectorsString: sel, OptionalSelectorsString: optSel, Group: group, Key: key, Output: outputFile})
	}

	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	log.Println("starting agent...")
	start(config)
	log.Println("initial configurations applied! update listening...")

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	stop()
	log.Println("Shutting down...")
}

var clients []*client.Client

func start(cfg *Config) {
	// group by selectors in order for client creation
	m := map[string][]struct {
		SelectorsString         string
		OptionalSelectorsString string
		Group                   string
		Key                     string
		Output                  string
	}{}
	for _, v := range cfg.ConfigList {
		key := fmt.Sprint(v.SelectorsString, "::", v.OptionalSelectorsString)
		m[key] = append(m[key], struct {
			SelectorsString         string
			OptionalSelectorsString string
			Group                   string
			Key                     string
			Output                  string
		}{
			SelectorsString:         v.SelectorsString,
			OptionalSelectorsString: v.OptionalSelectorsString,
			Group:                   v.Group,
			Key:                     v.Key,
			Output:                  v.Output,
		})
	}

	taskQueue := make(chan struct {
		Output string
		Val    []byte
	}, 1024) // use 1024 to retain enough pending writing items even when filesystem is failed to write for short period
	// create clients and add listeners
	for _, val := range m {
		sample := val[0]
		var sel = new(configapi.Selectors)
		var optsel = new(configapi.Selectors)
		if err := sel.Fill(sample.SelectorsString); err != nil {
			log.Fatal(err)
		}
		if err := optsel.Fill(sample.OptionalSelectorsString); err != nil {
			log.Fatal(err)
		}
		c := client.NewClient(srvList, client.ClientOptions{
			OverrideSelectors:         sel,
			OverrideOptionalSelectors: optsel,
		})
		for _, item := range val {
			output := item.Output
			c.AddConfigurationRequirement(client.RequiredConfig{
				Required: configapi.RequestedConfigurationKey{
					Group: item.Group,
					Key:   item.Key,
				},
				Callback: func(cfg configapi.Configuration) {
					select {
					case taskQueue <- struct {
						Output string
						Val    []byte
					}{Output: output, Val: cfg.Value}:
					default:
						log.Panicln(errors.New("task queue full and there should be errors processing configuration update"))
					}
				},
			})
		}
		clients = append(clients, c)
	}
	// start clients
	for _, c := range clients {
		if err := c.StartClient(); err != nil {
			log.Fatal(err)
		}
		if err := c.WaitStartupConfigureLoaded(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
	// start file writer
	go func() {
		for {
			item := <-taskQueue
			const maxRetry = 10
			for i := 0; i < maxRetry; i++ {
				f := func() error {
					fi, err := os.OpenFile(item.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						return err
					}
					defer func(fi *os.File) {
						_ = fi.Close()
					}(fi)
					if _, err := fi.Write(item.Val); err != nil {
						return err
					}
					return nil
				}
				if err := f(); err != nil {
					log.Println("Update failed! file:", item.Output, "error:", err)
					time.Sleep(1 * time.Second)
					continue
				} else {
					log.Println("Update success! file:", item.Output)
					break
				}
			}
		}
	}()
}

func stop() {
	var errs []error
	for _, c := range clients {
		if err := c.StopClient(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		log.Fatal(errors.Join(errs...))
	}
}
