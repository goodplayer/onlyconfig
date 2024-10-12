package client

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

func TestClientBasic(t *testing.T) {
	c := NewClient([]string{"http://127.0.0.1:8800"}, ClientOptions{
		SelectorDatacenter: "dc1",
	})

	c.AddConfigurationRequirement(RequiredConfig{
		Required: configapi.RequestedConfigurationKey{
			Group:   "group_json",
			Key:     "key_json",
			Version: "",
		},
		Callback: func(cfg configapi.Configuration) {
			t.Log(cfg)
		},
	})

	if err := c.StartClient(); err != nil {
		t.Fatal(err)
	}
	defer func(c *Client) {
		err := c.StopClient()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	if err := c.WaitStartupConfigureLoaded(context.Background()); err != nil {
		t.Fatal(err)
	}
}

type Container struct {
	Str  string `json:"str"`
	Int  int    `json:"int"`
	Bool bool   `json:"bool"`
}

func TestClientAdvBasic(t *testing.T) {
	c := NewClient([]string{"http://127.0.0.1:8800"}, ClientOptions{
		SelectorDatacenter: "dc1",
	})

	container := new(Container)
	var newContainer *atomic.Value
	ca := NewClientAdv(c)
	if nc, err := ca.RegisterJsonContainer("group_json", "key_json", container); err != nil {
		t.Fatal(err)
	} else {
		newContainer = nc
	}

	if err := c.StartClient(); err != nil {
		t.Fatal(err)
	}
	defer func(c *Client) {
		err := c.StopClient()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	if err := c.WaitStartupConfigureLoaded(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Log(*newContainer.Load().(*Container))
}
