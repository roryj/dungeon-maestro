package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
)

const (
	DICE_ROLL = "roll"
)

var actionList []string

type DndAction interface {
	ProcessAction() (string, error)
}

type Action struct {
	ActionType string `json:"action_type"`
}

type DiceRoll struct {
	DiceType string
	NumberOfDice int
}

type SlackRequest struct {
	Token string `json:"token"`
	TeamId string `json:"team_id"`
	TeamDomain string `json:"team_domain"`
	ChannelId string `json:"channel_id"`
	UserId string `json:"user_id"`
	UserName string `json:"user_name"`
	Command string `json:"command"`
	ResponseUrl string `json:"response_url"`
}

func (d *DiceRoll) ProcessAction() (string, error) {
	return "120", nil
}

func init() {
	fmt.Println("cold start")
	actionList = append(actionList, DICE_ROLL)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Handling request")

	fmt.Printf("Request: %v\n", request)
	fmt.Printf("body: %v\n", request.Body)
	//token=tnJiNSva3YKBc4mDkexP3LPt&team_id=TDWHRGTA6&team_domain=curseofstrahdtalk&channel_id=DDWHURXA6&channel_name=directmessage&user_id=UDY06M39C&
	// user_name=roryjacob&command=%2Froll&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTDWHRGTA6%2F477476469024%2FJCPrs6RD9awZmCpRsD1jCxNM&trigger_id=478111677203.472603571346.d7fe7114ecd1ffed30f481182f180912

	sr, err := parseSlackRequest(request.Body)
	if err != nil {
		fmt.Printf("Unable to marshal request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, errors.New("unable to parse request")
	}

	fmt.Printf("unmarshalled correctly! %v", sr)

	result := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body: "rolled a dice!",
	}
	return result, nil
}

// parseSlackRequest takes a slack request body with its crazy "&" splitting and attempts to turn it into
// a SlackRequest struct. The request looks something like:
// token=1234562423&team_id=sadjsakdjasd&team_domain=domain_team&channel_id=id&channel_name=directmessage&user_id=someuserid&user_name=somuser&command=%2Froll&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTDWHRGTA6%12341353%2FJCPrs6RD9awZmCpRsD1jCxNM&trigger_id=478111677203.472603571346.d7fe7114ecd1ffed30f481182f180912
func parseSlackRequest(request string) (SlackRequest, error) {

	fmt.Printf("Body: %s\n", request)

	p1 := strings.Replace(request, "=", "\": \"", 0)
	p2 := strings.Replace(p1, "&", "\", \"", 0)
	p3 := "{ \"" + p2 + "\"}"

	fmt.Printf("final payload: %s\n", p3)

	var sr SlackRequest
	err := json.Unmarshal([]byte(p3), &sr)
	return sr, err
}

func main() {
	fmt.Println("Main")
	lambda.Start(Handler)
}