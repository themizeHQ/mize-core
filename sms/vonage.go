package sms

import (
	"fmt"
	"os"

	"github.com/vonage/vonage-go-sdk"
	"go.uber.org/zap"
	"mize.app/logger"
)

func SendSms(to string, message string) error {
	auth := vonage.CreateAuthFromKeySecret(os.Getenv("VONAGE_API_KEY"), os.Getenv("VONAGE_API_SECRET"))
	smsClient := vonage.NewSMSClient(auth)
	response, smsErr, err := smsClient.Send("Mize", to, message, vonage.SMSOpts{})
	if response.Messages[0].Status != "0" {
		logger.Error(fmt.Errorf("vonage - failed to send sms to %s", to), zap.Error(err), zap.Any("sms error", smsErr))
		return err
	}
	logger.Info(fmt.Sprintf("vonage - sms sent to %s", to))
	return nil
}
