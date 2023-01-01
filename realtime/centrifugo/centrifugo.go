package centrifugo

import (
	"os"

	"mize.app/network"
)

type CentrifugoController struct {
	BaseUrl string
}

func (c *CentrifugoController) Publish(channel string, scope string, data interface{}) error {
	network := network.NetworkController{BaseUrl: c.BaseUrl}
	_, err := network.Post("", &map[string]string{
		"Authorization": os.Getenv("CENTRIFUGO_API_KEY"),
	}, &map[string]interface{}{
		"params": map[string]interface{}{
			"data": map[string]interface{}{
				"payload": data,
				"scope":   scope,
			},
			"channel": channel,
		},
	}, nil)
	if err != nil {
		return err
	}
	return nil
}
