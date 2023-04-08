package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/atuprosper/booking-project/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	// Go routine function that runs in the background
	go func() {
		for {
			message := <-app.MailChannel
			sendMessage(message)
		}
	}()
}

func sendMessage(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// ioutil is used to read files
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-template/%s", m.Template))

		if err != nil {
			app.ErrorLog.Println(err)
		}

		// convert the []byte returned to string
		mailTemplate := string(data)

		// Replace the template litrals [%body%] with the content passed
		messageToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, messageToSend)
	}

	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}

}
