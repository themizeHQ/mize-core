package centrifugo

import (
	"encoding/json"
	"fmt"
	"os"

	"mize.app/network"
)

type CentrifugoController struct {
	BaseUrl string
}

func (c *CentrifugoController) Publish(channel string, scope string, data interface{}) error {
	fmt.Println(channel)
	network := network.NetworkController{BaseUrl: c.BaseUrl}
	res, err := network.Post("", &map[string]string{
		"Authorization": fmt.Sprintf("apikey %s", os.Getenv("CENTRIFUGO_API_KEY")),
	}, &map[string]interface{}{
		"method": "publish",
		"params": map[string]interface{}{
			"data": map[string]interface{}{
				"payload": data,
				"scope":   scope,
			},
			"channel": channel,
		},
	}, nil)
	var t map[string]interface{}
	json.Unmarshal([]byte(*res), &t)
	if err != nil {
		return err
	}
	return nil
}
