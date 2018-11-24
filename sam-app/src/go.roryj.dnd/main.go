package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"go.roryj.dnd/slack"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	diceRoll      = "roll"
	identifySpell = "spell"
)

const slackUrlEndpointFormat = "https://hooks.slack.com/services/%s"
var slackWebhookEndpoint string

var stage string
var region string

func init() {
	stage = os.Getenv("Stage")
	region = os.Getenv("Region")

	log.Printf("starting invoke in %s and %s", stage, region)

	secrets, err := getSecret()
	if err != nil {
		log.Printf("failed to get secret from ssm: %v", err)
		log.Printf("no SlackWebhookPath env var set. Not sending update to webhook")
	} else {
		log.Printf("able to retrieve secrets")
		if secrets.SlackWebhookUrl == "" {
			log.Printf("no secret for the slack webhook url")
		} else {
			slackWebhookEndpoint = fmt.Sprintf(slackUrlEndpointFormat, secrets.SlackWebhookUrl)
		}
	}
}

type lambdaSecrets struct {
	SlackWebhookUrl string `json:"SlackWebhookUrl"`
}

func getSecret() (lambdaSecrets, error) {
	secretName := "beta/dungeonMaestro/slack"

	//Create a Secrets Manager client
	session.NewSession()
	sess := session.Must(session.NewSession())
	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return lambdaSecrets{}, err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secrets lambdaSecrets
	err = json.Unmarshal([]byte(*result.SecretString), &secrets)

	return secrets, err
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sr, err := parseSlackRequest(request.Body)
	if err != nil {
		fmt.Printf("Unable to marshal request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
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
				Body:       fmt.Sprintf("%s", err),
			}, nil
		}

		break
	case identifySpell:
		log.Printf("Detected identify spell request.")
		action, err = NewIdentifySpell(sr.Text)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       fmt.Sprintf("%s", err),
			}, nil
		}
	default:
		log.Printf("Unable to determine request type: %s", command)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "unknown request type. Only [/roll, /spell] are accepted",
		}, nil
	}

	actionResult, err := action.ProcessAction()
	if err != nil {
		if v, ok := err.(*DndActionError); ok {
			if v.GetType() == UserError {
				return events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body: v.message,
				}, nil
			} else {
				log.Printf("error processing action %v. %v", action, err)
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, errors.New("unable to parse request")
			}
		} else {
			log.Printf("the error type from ProcessAction was an unexpected type: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
			}, errors.New("internal error")
		}
	}

	postSlackUpdate(actionResult)
	result := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       actionResult.Text,
	}

	return result, nil
}

func postSlackUpdate(result slack.WebhookResponse) error {
	if slackWebhookEndpoint == "" {
		log.Printf("no endpoint set, not sending slack update")
		return nil
	}

	b, err := json.Marshal(result)
	if err !=  nil {
		return err
	}

	_, err = http.Post(slackWebhookEndpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		return  err
	}

	log.Printf("successfully posted to slack")

	return nil
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
