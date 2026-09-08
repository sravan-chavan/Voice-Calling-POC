package main

import "fmt"
import "github.com/plivo/plivo-go/v7"

func main() {
	client, err := plivo.NewClient("<auth_id>","<auth_token>", &plivo.ClientOptions{})
	if err != nil {
			fmt.Print("Error", err.Error())
			return
		}
	response, err := client.Calls.Create(
		plivo.CallCreateParams{
			From: "<caller_id>",
			To: "destination_number1<destination_number2",
			AnswerURL: "https://s3.amazonaws.com/static.plivo.com/answer.xml",
			AnswerMethod: "GET",
		},
	)
	if err != nil {
			fmt.Print("Error", err.Error())
			return
		}
	fmt.Printf("Response: %#v\n", response)
}