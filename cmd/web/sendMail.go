package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	result, resp, err := sib.AccountApi.GetAccount(ctx)
	if err != nil {
		fmt.Println("Error when calling AccountApi->get_account: ", err.Error())
		return
	}
	fmt.Println("GetAccount Object:", result, " GetAccount Response: ", resp)

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
		TextContent: m.Content,
	}

	// Send the email using the Sendinblue API
	created, response, err := sib.TransactionalEmailsApi.SendTransacEmail(ctx, message)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("created:", created)
	fmt.Println("response:", response)

	// Print a message indicating that the email was sent successfully
	fmt.Println("Email sent successfully!")

}
