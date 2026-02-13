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

	peClient, err := sdk.New(sdk.Config{
		AccessToken: "YOUR_ACCESS_TOKEN",
		Country:     "PE",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Peru (PE) ===")
	showCapabilities(ctx, peClient)

	mxClient, err := peClient.ForCountry("MX")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Mexico (MX) ===")
	showCapabilities(ctx, mxClient)

	fmt.Println("\n=== All Supported Regions ===")
	regions, err := peClient.Capabilities.ListRegions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range regions {
		fmt.Printf("- %s (%s) - Currency: %s\n", r.CountryCode, r.Locale, r.CurrencyCode)
	}

	fmt.Println("\n=== Cross-Region Comparison ===")
	compareRegions(ctx, peClient)
}

func showCapabilities(ctx context.Context, client *sdk.SDK) {
	caps, err := client.Capabilities.Get(ctx)
	if err != nil {
		log.Printf("Error getting capabilities: %v", err)
		return
	}

	fmt.Printf("Country: %s\n", caps.Region.CountryCode)
	fmt.Printf("Currency: %s\n", caps.Region.CurrencyCode)
	fmt.Printf("Timezone: %s\n", caps.Region.TimezoneIANA)

	fmt.Println("\nPayment Methods:")
	for _, m := range caps.Payment.SupportedMethods {
		fmt.Printf("  - %s (%s): %.2f - %.2f %s\n",
			m.Name, m.Type, m.MinAmount.Amount, m.MaxAmount.Amount, m.MinAmount.Currency)
	}

	fmt.Printf("\nMax Installments: %d\n", caps.Payment.MaxInstallments)
	fmt.Printf("Supports Refunds: %v\n", caps.Payment.SupportsRefunds)

	fmt.Println("\nCarriers:")
	for _, c := range caps.Shipment.SupportedCarriers {
		fmt.Printf("  - %s: %v\n", c.Name, c.ServiceTypes)
	}

	fmt.Printf("\nQR Supported: %v\n", caps.QR.Supported)
	if caps.QR.Supported {
		fmt.Printf("  Dynamic QR: %v\n", caps.QR.SupportsDynamicQR)
		fmt.Printf("  Static QR: %v\n", caps.QR.SupportsStaticQR)
		fmt.Printf("  Max Amount: %.2f %s\n", caps.QR.MaxAmount.Amount, caps.QR.MaxAmount.Currency)
	}
}

func compareRegions(ctx context.Context, client *sdk.SDK) {
	countries := []string{"PE", "MX", "AR", "BR"}

	fmt.Printf("%-10s %-10s %-15s %-10s\n", "Country", "Currency", "Max Amount", "Installments")
	fmt.Println("----------------------------------------------------")

	for _, country := range countries {
		caps, err := client.Capabilities.GetForCountry(ctx, country)
		if err != nil {
			continue
		}

		fmt.Printf("%-10s %-10s %-15.2f %-10d\n",
			caps.Region.CountryCode,
			caps.Region.CurrencyCode,
			caps.Payment.MaxAmount.Amount,
			caps.Payment.MaxInstallments)
	}
}

func exampleValidation(ctx context.Context, client *sdk.SDK) {
	payment, err := client.Payment.Create(ctx, &domain.CreatePaymentRequest{
		ExternalReference: "order-12345",
		Amount: domain.Money{
			Amount:   100.00,
			Currency: "PEN",
		},
		Description: "Test payment",
		Payer: domain.Payer{
			Email: "test@example.com",
		},
	})
	if err != nil {
		log.Printf("Validation or creation error: %v", err)
		return
	}

	fmt.Printf("Payment created: %s\n", payment.ID)
}
