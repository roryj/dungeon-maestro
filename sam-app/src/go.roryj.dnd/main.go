package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"strings"
)

const (
	diceRoll = "roll"
	identifySpell = "spell"
)



type SlackRequest struct {
	Token string `json:"token"`
	TeamId string `json:"team_id"`
	Text string `json:"text"`
	TeamDomain string `json:"team_domain"`
	ChannelId string `json:"channel_id"`
	UserId string `json:"user_id"`
	UserName string `json:"user_name"`
	Command string `json:"command"`
	ResponseUrl string `json:"response_url"`
}

func init() {
	fmt.Println("cold start")
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Handling request")

	log.Printf("Request: %v\n", request)
	log.Printf("body: %v\n", request.Body)

	sr, err := parseSlackRequest(request.Body)
	if err != nil {
		fmt.Printf("Unable to marshal request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
		}, err
	}

	var action DndAction

	// remove html encoded "/" character from the front of the command
	command := strings.Replace(sr.Command, "%2F", "", 1)
	log.Printf("processing command %s", command)


	switch command {
	case diceRoll: // for a dice roll, we expect the following format: /roll <number-of-dice> d<dice-type> ie. /roll 10 d4
		log.Printf("Detected dice roll request.")
		action, err = NewDiceRoll(sr.Text)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body: fmt.Sprintf("%s", err),
			}, nil
		}

		break
	case identifySpell:
		log.Printf("Detected identify spell request.")
		action, err = NewIdentifySpell(sr.Text)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body: fmt.Sprintf("%s", err),
			}, nil
		}
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body: "unknown request type. Only [/roll /spell] are accepted",
		}, nil
	}

	actionResult, err := action.ProcessAction()
	if err != nil {
		log.Fatalf("error processing action %v. %v", action, err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, errors.New("unable to parse request")
	}

	result := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body: actionResult,
	}

	return result, nil
}

// parseSlackRequest takes a slack request body with its crazy "&" splitting and attempts to turn it into
// a SlackRequest struct. The request looks something like:
// token=1234562423&team_id=sadjsakdjasd&team_domain=domain_team&channel_id=id&channel_name=directmessage&user_id=someuserid&user_name=somuser&command=%2Froll&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTDWHRGTA6%12341353%2FJCPrs6RD9awZmCpRsD1jCxNM&trigger_id=478111677203.472603571346.d7fe7114ecd1ffed30f481182f180912
func parseSlackRequest(request string) (SlackRequest, error) {
	p1 := strings.Replace(request, "=", "\": \"", -1)
	p2 := strings.Replace(p1, "&", "\", \"", -1)
	p3 := "{ \"" + p2 + "\"}"

	var sr SlackRequest
	err := json.Unmarshal([]byte(p3), &sr)
	return sr, err
}

func main() {
	lambda.Start(Handler)
}