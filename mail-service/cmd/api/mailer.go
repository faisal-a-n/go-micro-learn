package main

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (this *Mail) SendSMTPMessage(message Message) error {
	//Default from address
	if message.From == "" {
		message.From = this.FromAddress
	}
	if message.FromName == "" {
		message.FromName = this.FromName
	}

	data := map[string]any{
		"message": message.Data,
	}

	message.DataMap = data

	//Build messages
	formattedMessage, err := this.BuildHTMLMessage(message)
	if err != nil {
		return err
	}

	plainMessage, err := this.BuildPlainTextMessage(message)
	if err != nil {
		return err
	}

	//Setup SMTP client
	server := mail.NewSMTPClient()
	server.Host = this.Host
	server.Port = this.Port
	server.Username = this.Username
	server.Password = this.Password
	server.Encryption = this.getEncryption(this.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}
	log.Println("Creating mail", message)
	//Setup email
	email := mail.NewMSG()
	email.SetFrom(message.From).
		AddTo(message.To).
		SetSubject(message.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(message.Attachments) > 0 {
		for _, v := range message.Attachments {
			email.AddAttachment(v)
		}
	}

	//Send mail
	err = email.Send(smtpClient)
	if err != nil {
		return err
	}
	return nil
}

func (this *Mail) BuildHTMLMessage(message Message) (string, error) {
	templateToRender := "./templates/mail.html.gohtml"
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", nil
	}

	formattedMessage := tpl.String()
	formattedMessage, err = this.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (this *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

func (this *Mail) BuildPlainTextMessage(message Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"
	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", nil
	}

	plainMessage := tpl.String()
	plainMessage, err = this.inlineCSS(plainMessage)
	if err != nil {
		return "", err
	}

	return plainMessage, nil
}

func (this *Mail) getEncryption(encryption string) mail.Encryption {
	switch encryption {
	case "tls":
		return mail.EncryptionTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
