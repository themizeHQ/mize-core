package sms

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
	"mize.app/logger"
	"mize.app/network"
)

type TermiiService struct {
	BaseUrl string
}

func (t *TermiiService) SendSms(to string, message string) error {
	network := network.NetworkController{BaseUrl: t.BaseUrl}
	r, err := network.Post("/api/sms/send", nil, &map[string]interface{}{
		"api_key": os.Getenv("TERMII_API_KEY"),
		"to":      to,
		"from":    "Mize HQ",
		"sms":     message,
		"type":    "plain",
		"channel": "generic",
	}, nil)
	if err != nil {
		logger.Error(fmt.Errorf("termii - failed to send sms to %s", to), zap.Error(err))
		return err
	}
	var res map[string]interface{}
	err = json.Unmarshal([]byte(*r), &res)
	if err != nil {
		logger.Error(fmt.Errorf("termii - failed to send sms to %s", to), zap.Error(err))
		return err
	}
	if res["code"] != "ok" {
		err = fmt.Errorf("termii - failed to send sms to %s", to)
		logger.Error(err, zap.Any("response", res))
		return err
	}
	logger.Info(fmt.Sprintf("termii - sms sent to %s", to))
	return nil
}
