package emails

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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
		log.Println(err)
		return false
	} else {
		fmt.Println(response.Body)
		fmt.Println(response.StatusCode)
		if response.StatusCode != 200 {
			return false
		}
		return true
	}
}

func loadTemplates(templateName string, opts interface{}) bytes.Buffer {
	var buffer bytes.Buffer
	template.Must(template.ParseFiles(fmt.Sprintf(filepath.Join(dir, "/templates/emails/%s.html"), templateName))).Execute(&buffer, opts)
	return buffer
}
