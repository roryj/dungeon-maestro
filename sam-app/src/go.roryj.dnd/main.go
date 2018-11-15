package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

const (
	diceRoll = "roll"
)

type DndAction interface {
	ProcessAction() (string, error)
}

type Action struct {
	ActionType string `json:"action_type"`
}

type DiceRoll struct {
	DiceSides int
	NumberOfDice int
}

// NewDiceRoll takes the input text for the /roll command and turns it into a DiceRoll struct
// The format for the dice roll command is /roll <num-dice> d<dice-type>
// ex: /roll 10 d20
func NewDiceRoll(input string) (*DiceRoll, error) {
	// split on html encoded spaces (+)
	split := strings.Split(input, "+")

	if len(split) != 2 {
		return &DiceRoll{}, fmt.Errorf("incorrect number of dice roll parameters. ex: /roll 12 d20")
	}

	// first item should be an int
	numDice, err := strconv.Atoi(split[0])
	if err != nil {
		return &DiceRoll{}, fmt.Errorf("the first parameter, %s, needs to be the number of dice to roll (ex: 10)", split[0])
	}

	// second item should start with the letter d, and followed by an integer
	if !strings.HasPrefix(split[1], "d") {
		return &DiceRoll{}, fmt.Errorf("the second parameter, %s, needs to be the number of sides on the dice (ex: d20)", split[0])
	}

	dS := strings.Replace(split[1], "d", "", 1)

	diceSides, err := strconv.Atoi(dS)
	if err != nil {
		return &DiceRoll{}, fmt.Errorf("the second parameter, %s, needs to be the number of sides on the dice (ex: d20)", split[0])
	}

	return &DiceRoll{
		DiceSides: diceSides,
		NumberOfDice: numDice,
	}, nil
}

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

func (d *DiceRoll) ProcessAction() (string, error) {

	var total int

	for i := 0; i < d.NumberOfDice; i++ {
		total += rand.Intn(d.DiceSides) + 1
	}

	return fmt.Sprintf("Rolled %d d%d and got %d\n", d.NumberOfDice, d.DiceSides, total), nil
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

	log.Printf("processing command %s", sr.Command)

	// remove html encoded "/" character from the front of the command
	command := strings.Replace(sr.Command, "%2F", "", 1)

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
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body: "unknown request type. Only [/roll] is accepted",
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

	log.Printf("This is the result: %s", p3)

	var sr SlackRequest
	err := json.Unmarshal([]byte(p3), &sr)
	return sr, err
}

func main() {
	lambda.Start(Handler)
}