package sms

type SmsServiceType interface {
	SendSms(to string, message string) error
}
