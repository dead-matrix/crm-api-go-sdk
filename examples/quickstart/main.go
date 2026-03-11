package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi"
)

func main() {
	client, err := crmapi.NewClient(crmapi.Config{
		BaseURL:      "https://your-crm.example",
		StaffID:      123,
		ServiceToken: "YOUR_SERVICE_TOKEN",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	user, err := client.GetUser(ctx, 7014133383)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", user)
}
