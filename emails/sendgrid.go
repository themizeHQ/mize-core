package emails

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
	"mize.app/logger"
)

var dir, _ = os.Getwd()

func SendEmail(toEmail string, subject string, templateName string, opts interface{}) bool {
	from := mail.NewEmail("Emeka from Mize", os.Getenv("MIZE_EMAIL"))
	to := mail.NewEmail("Example User", toEmail)
	buffer := loadTemplates(templateName, opts)
	message := mail.NewSingleEmail(from, subject, to, "", buffer.String())
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		logger.Error(err, zap.String("to", toEmail), zap.String("template_name", templateName))
		return false
	} else {
		if response.StatusCode == 202 {
			logger.Info("request completed without errors", zap.Any("response", response))
			return true
		} else {
			logger.Info("failed to complete request", zap.Any("response", response))
			return false
		}
	}
}

func loadTemplates(templateName string, opts interface{}) bytes.Buffer {
	var buffer bytes.Buffer
	template.Must(template.ParseFiles(fmt.Sprintf(filepath.Join(dir, "/templates/emails/%s.html"), templateName))).Execute(&buffer, opts)
	return buffer
}
