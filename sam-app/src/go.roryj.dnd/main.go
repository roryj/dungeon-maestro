package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.roryj.dnd/slack"
	"log"
	"os"
	"strings"
)

const (
	diceRoll      = "roll"
	identifySpell = "spell"
)

var stage string
var region string

func init() {
	stage = os.Getenv("Stage")
	region = os.Getenv("Region")
	log.Printf("starting container with stage: %s region: %s", stage, region)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sr, err := parseSlackRequest(request.Body)
	if err != nil {
		log.Printf("Unable to marshal request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       err.Error(),
		}, err
	}

	var action DndAction

	// remove html encoded "/" character from the front of the command
	command := strings.Replace(sr.Command, "%2F", "", 1)
	log.Printf("Processing command %s", command)

	switch command {
	case diceRoll: // for a dice roll, we expect the following format: /roll <number-of-dice> d<dice-type> ie. /roll 10 d4
		log.Printf("Detected dice roll request.")
		action, err = NewDiceRoll(sr.UserName, sr.Text)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       err.Error(),
			}, nil
		}

		break
	case identifySpell:
		log.Printf("Detected identify spell request.")
		action, err = NewIdentifySpell(sr.Text)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       err.Error(),
			}, nil
		}
	default:
		log.Printf("Unable to determine request type: %s", command)
		return events.APIGatewayProxyResponse{
			StatusCode: 200, // return a 200 so that the user sees the response
			Body:       "unknown request type. Only [/roll, /spell] are accepted",
		}, nil
	}

	actionResult, err := action.ProcessAction()
	if err != nil {
		if v, ok := err.(*DndActionError); ok {
			if v.GetType() == UserError {
				return events.APIGatewayProxyResponse{
					StatusCode: 200, // return a 200 so that the result is seen by the user
					Body:       v.message,
				}, nil
			} else {
				log.Printf("error processing action: %v. %v", action, err)
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, errors.New("service error")
			}
		} else {
			log.Printf("the error type from ProcessAction was an unexpected type: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
			}, errors.New("service error")
		}
	}

	r, err := json.Marshal(actionResult)
	if err != nil {
		log.Printf("failed to jsonify response payload: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, errors.New("service error")
	}

	result := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(r),
	}

	return result, nil
}

// parseSlackRequest takes a slack request body with its crazy "&" splitting and attempts to turn it into
// a Request struct. The request looks something like:
// token=1234562423&team_id=sadjsakdjasd&team_domain=domain_team&channel_id=id&channel_name=directmessage&user_id=someuserid&user_name=somuser&command=%2Froll&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTDWHRGTA6%12341353%2FJCPrs6RD9awZmCpRsD1jCxNM&trigger_id=478111677203.472603571346.d7fe7114ecd1ffed30f481182f180912
func parseSlackRequest(request string) (slack.Request, error) {
	p1 := strings.Replace(request, "=", "\": \"", -1)
	p2 := strings.Replace(p1, "&", "\", \"", -1)
	p3 := "{ \"" + p2 + "\"}"

	var sr slack.Request
	err := json.Unmarshal([]byte(p3), &sr)
	return sr, err
}

func main() {
	lambda.Start(Handler)
}
