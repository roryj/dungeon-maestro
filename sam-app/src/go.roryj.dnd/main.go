package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	fmt.Println("cold start")
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Handling request")
	result := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body: "rolled a dice!",
	}
	return result, nil
}

func main() {
	fmt.Println("Main")
	lambda.Start(Handler)
}