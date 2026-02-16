package main

import (
	"context"
	"fmt"
	"log"

	sdk "github.com/zentry/sdk-mercadolibre"
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
		fmt.Printf("- %s (%s) Currency: %s\n", r.CountryCode, r.Locale, r.CurrencyCode)
	}

	fmt.Println("\n=== Cross-Region Comparison ===")
	countries := []string{"PE", "MX", "AR", "BR"}
	fmt.Printf("%-10s %-10s %-15s %-10s\n", "Country", "Currency", "Max Amount", "Installments")
	for _, country := range countries {
		caps, err := peClient.Capabilities.GetForCountry(ctx, country)
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

func showCapabilities(ctx context.Context, client *sdk.SDK) {
	caps, err := client.Capabilities.Get(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Country: %s\n", caps.Region.CountryCode)
	fmt.Printf("Currency: %s\n", caps.Region.CurrencyCode)

	fmt.Println("Payment Methods:")
	for _, m := range caps.Payment.SupportedMethods {
		fmt.Printf("  - %s (%s): %.2f - %.2f %s\n",
			m.Name, m.Type, m.MinAmount.Amount, m.MaxAmount.Amount, m.MinAmount.Currency)
	}

	fmt.Printf("Max Installments: %d\n", caps.Payment.MaxInstallments)
	fmt.Printf("QR Supported: %v\n", caps.QR.Supported)
}
