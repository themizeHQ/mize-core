package sms

import (
	"fmt"
	"os"

	"github.com/vonage/vonage-go-sdk"
)

func SendSms(to string, message string) error {
	auth := vonage.CreateAuthFromKeySecret(os.Getenv("VONAGE_API_KEY"), os.Getenv("VONAGE_API_SECRET"))
	smsClient := vonage.NewSMSClient(auth)
	response, smsErr, err := smsClient.Send("Mize", to, message, vonage.SMSOpts{})
	if response.Messages[0].Status != "0" {
		fmt.Println(err)
		fmt.Println(smsErr)
		return err
	}
	return nil
}
