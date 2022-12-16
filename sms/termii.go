package sms

import (
	"fmt"
	"os"

	"mize.app/logger"
	"mize.app/network"
)

type TermiiService struct {
	BaseUrl string
}

func (t *TermiiService) SendSms(to string, message string) error {
	network := network.NetworkController{BaseUrl: t.BaseUrl}
	_, err := network.Post("/api/sms/send", nil, &map[string]interface{}{
		"api_key": os.Getenv("TERMII_API_KEY"),
		"to":      to,
		"from":    "Mize HQ",
		"sms":     message,
		"type":    "plain",
		"channel": "generic",
	}, nil)
	if err != nil {
		logger.Info(fmt.Sprintf("termii - failed to send sms to %s", to))
		return err
	}
	logger.Info(fmt.Sprintf("termii - sms sent to %s", to))
	return nil
}
