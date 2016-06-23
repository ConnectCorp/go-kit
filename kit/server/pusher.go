package server

import (
	"github.com/pusher/pusher-http-go"
)

// MustInitPusher initializes the Pusher client, or panics.
func MustInitPusher(commonCfg *CommonConfig, pusherCfg *PusherConfig) *pusher.Client {
	pusherClient, err := pusher.ClientFromURL(pusherCfg.PusherSpec)
	if err != nil {
		panic(err)
	}
	pusherClient.HttpClient = MakeHTTPClientForConfig(commonCfg)
	return pusherClient
}
