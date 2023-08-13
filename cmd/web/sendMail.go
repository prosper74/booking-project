package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/atuprosper/booking-project/internal/models"
	"github.com/joho/godotenv"
	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
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
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	SENDINBLUE_API_KEY := os.Getenv("SENDINBLUE_API_KEY")

	var ctx context.Context
	cfg := sendinblue.NewConfiguration()
	//Configure API key authorization: api-key
	cfg.AddDefaultHeader("api-key", SENDINBLUE_API_KEY)
	//Configure API key authorization: partner-key
	cfg.AddDefaultHeader("partner-key", SENDINBLUE_API_KEY)

	sib := sendinblue.NewAPIClient(cfg)
	_, _, err = sib.AccountApi.GetAccount(ctx)
	if err != nil {
		fmt.Println("Error when calling AccountApi->get_account: ", err.Error())
		return
	}

	var emailContent string

	if m.Template == "" {
		emailContent = m.Content
	} else {
		// ioutil is used to read files
		data, err := os.ReadFile(fmt.Sprintf("./email-template/%s", m.Template))

		if err != nil {
			app.ErrorLog.Println(err)
		}

		// convert the []byte returned to string
		mailTemplate := string(data)

		// Replace the template litrals [%body%] with the content passed
		messageToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		emailContent = messageToSend
	}

	// Create an email message
	message := sendinblue.SendSmtpEmail{
		Sender: &sendinblue.SendSmtpEmailSender{
			Name:  "Hotel Bookings",
			Email: m.From,
		},
		To: []sendinblue.SendSmtpEmailTo{
			{
				Email: m.To,
				Name:  "",
			},
		},
		Subject:     m.Subject,
		TextContent: emailContent,
	}

	// Send the email using the Sendinblue API
	_, _, err = sib.TransactionalEmailsApi.SendTransacEmail(ctx, message)
	if err != nil {
		log.Println(err)
		return
	}

	// Print a message indicating that the email was sent successfully
	fmt.Println("Email sent successfully!")

}
