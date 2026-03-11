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

	username := "johndoe"

	input := crmapi.CreateUserInput{
		UserID:   7014133383,
		FullName: "John Doe",
		Username: &username,
		BotID:    1,
	}

	res, err := client.CreateUser(context.Background(), input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", res)
}
