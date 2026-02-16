package main

import (
	"context"
	"fmt"
	"log"

	sdk "github.com/zentry/sdk-mercadolibre"
	"github.com/zentry/sdk-mercadolibre/core/domain"
)

func main() {
	ctx := context.Background()

	client, err := sdk.New(sdk.Config{
		AccessToken: "YOUR_ACCESS_TOKEN",
		Country:     "PE",
	})
	if err != nil {
		log.Fatal(err)
	}

	payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
		ExternalReference: "order-12345",
		Amount: domain.Money{
			Amount:   100.00,
			Currency: "PEN",
		},
		Description: "Test payment",
		Method:      domain.PaymentMethodCard,
		Payer: domain.Payer{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Payment created: %s\n", payment.ID)
	fmt.Printf("Status: %s\n", payment.Status.String())
}
