package client

import "github.com/meidoworks/nekoq-component/configure/configclient"

type ClientOptions = configclient.ClientOptions

type Client = configclient.Client

type RequiredConfig = configclient.RequiredConfig

func NewClient(serverList []string, opt ClientOptions) *Client {
	return configclient.NewClient(serverList, opt)
}

type ClientAdv = configclient.ClientAdv

type Unmarshaler = configclient.Unmarshaler

func NewClientAdv(c *Client) *ClientAdv {
	return configclient.NewClientAdv(c)
}
