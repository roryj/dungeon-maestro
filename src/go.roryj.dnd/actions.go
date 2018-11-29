package main

import (
	"fmt"
	"go.roryj.dnd/slack"
	"golang.org/x/net/html"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

const (
	UserError    = "USER"
	ServiceError = "SERVICE"
)

type DndActionError struct {
	message   string
	errorType string
}

func (e *DndActionError) Error() string {
	return e.message
}

func (e *DndActionError) GetType() string {
	return e.errorType
}

type DndAction interface {
	ProcessAction() (slack.WebhookResponse, error)
}

type DiceRoll struct {
	user         string
	diceSides    int
	numberOfDice int
}

// NewDiceRoll takes the input text for the /roll command and turns it into a DiceRoll struct
// The format for the dice roll command is /roll <num-dice> d<dice-type>
// ex: /roll 10 d20
func NewDiceRoll(user, input string) (*DiceRoll, error) {
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
		user:         user,
		diceSides:    diceSides,
		numberOfDice: numDice,
	}, nil
}

func (d *DiceRoll) ProcessAction() (slack.WebhookResponse, error) {
	log.Printf("Processing dice roll on %v", d)
	var total int

	for i := 0; i < d.numberOfDice; i++ {
		total += rand.Intn(d.diceSides) + 1
	}

	log.Printf("successfully rolled %d d%d and got %d", d.numberOfDice, d.diceSides, total)

	return slack.WebhookResponse{
		Text:         fmt.Sprintf("%s rolled %d d%d and got %d\n", d.user, d.numberOfDice, d.diceSides, total),
		ResponseType: slack.ShowResponseToAll,
	}, nil
}

const dndSpellEndoint = "https://www.dndbeyond.com/spells/"

type IdentifySpell struct {
	spellName  string
	statBlocks []string
}

func NewIdentifySpell(input string) (*IdentifySpell, error) {
	s := strings.Replace(input, "+", " ", -1)

	return &IdentifySpell{
		spellName:  s,
		statBlocks: []string{},
	}, nil
}

// IdentifySpell#ProcessAction gets info about a spell from internet sources. All spaces and "/" in spell names should be
// replaced by "-". Any apostrophes should be removed completely
// ex: 	Antimagic Field -> /spells/antimagic-field
// 		Antipathy/Sympathy -> /spells/antipathy-sympathy
func (s *IdentifySpell) ProcessAction() (slack.WebhookResponse, error) {
	log.Printf("Processing spell identify on %v", s)

	urlEncodedSpell := strings.Replace(s.spellName, " ", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "/", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "'", "", -1)
	urlEncodedSpell = strings.ToLower(urlEncodedSpell)

	fullPath := dndSpellEndoint + urlEncodedSpell

	log.Printf("url: %s", fullPath)

	resp, err := http.Get(fullPath)
	if err != nil {

		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("failed to make a request to dnd beyond. %v", err),
			errorType: ServiceError,
		}
	}
	defer resp.Body.Close()

	rootDoc, err := html.Parse(resp.Body)
	if err != nil {
		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("failed to parse the request from dnd beyod. %v", err),
			errorType: ServiceError,
		}
	}

	// first check to see if this is a 404 page (unknown spell)
	_, ok := findHtmlElement(rootDoc, "error-page error-page-404")
	if ok {
		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("unknown spell: %s", s.spellName),
			errorType: UserError,
		}
	}

	var missingAttributes []string
	spellAttributes := []string{"Level", "Casting Time", "Range/Area", "Components", "Duration", "School",
		"Attack/Save", "Damage/Effect"}
	currRow := 0
	maxInRow := 4

	// get the number of separate rows to display in slack based on the number of spell attributes, and the max number
	// to display in a row. We use separate attachments to identify rows for slack, and separate fields within an
	// attachment to specify columns
	numRows := int(math.Ceil(float64(len(spellAttributes) / maxInRow)))
	attachments := make([]slack.WebhookResponseAttachment, numRows)

	for index, a := range spellAttributes {
		log.Printf("Getting %s attribute", a)
		r, ok := getValueForAttribute(rootDoc, a)
		if !ok {
			missingAttributes = append(missingAttributes, a)
		} else {
			// this is horribly inefficient, and should most def be using pointers here, but this works, and is totally
			// fine for the most part
			attachment := attachments[currRow]
			fields := attachment.Fields
			fields = append(fields, slack.WebhookResponseAttachmentField{
				Title: a,
				Value: r,
				Short: true,
			})

			attachments[currRow] = slack.WebhookResponseAttachment{
				Fields: fields,
			}

			// if the next increment will cause us to be in the max per-row, increment the row count
			if index%maxInRow+1 == maxInRow {
				currRow++
			}
		}
	}

	if len(missingAttributes) > 0 {
		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("failed to find attributes for: %v", missingAttributes),
			errorType: ServiceError,
		}
	}

	return slack.WebhookResponse{
		Text:         fmt.Sprintf("Description of the spell %s", s.spellName),
		Attachments:  attachments,
		ResponseType: slack.ShowResponseToAll,
	}, nil
}

const classFormat = "ddb-statblock-item ddb-statblock-item-%s"
const valueClass = "ddb-statblock-item-value"

func getValueForAttribute(n *html.Node, dndAttribute string) (string, bool) {
	a := strings.ToLower(dndAttribute)
	a = strings.Replace(a, "/", "-", -1)
	a = strings.Replace(a, " ", "-", -1)
	attrClassName := fmt.Sprintf(classFormat, a)
	log.Printf("formatted class name: %s", attrClassName)

	attrNode, ok := findHtmlElement(n, attrClassName)
	if !ok {
		log.Printf("failed to find dnd attribute for %s", dndAttribute)
		return "", false
	}

	value, ok := findHtmlElement(attrNode, valueClass)
	if !ok {
		log.Printf("failed to find value for attribute %s", dndAttribute)
		return "", false
	}

	var result string

	// Most dnd attributes can be found as the text under ddb-statblock-item-value class for the attribute. However,
	// Damage/Effect and Components are different.
	switch dndAttribute {
	case "Attack/Save":
		result = strings.TrimSpace(value.LastChild.PrevSibling.FirstChild.Data)
	case "Damage/Effect":
		result = strings.TrimSpace(value.LastChild.Data)
	case "Components":
		result = strings.TrimSpace(value.FirstChild.NextSibling.FirstChild.Data)
	default:
		result = strings.TrimSpace(value.FirstChild.Data)
	}

	return result, true
}

func findHtmlElement(n *html.Node, className string) (*html.Node, bool) {

	if checkClassName(n, className) {
		return n, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := findHtmlElement(c, className)
		if result != nil {
			return result, ok
		}
	}

	return nil, false
}

func checkClassName(n *html.Node, className string) bool {
	a, ok := getAttribute(n, "class")
	if ok && a == className {
		return true
	}
	return false
}

func getAttribute(n *html.Node, attrKey string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == attrKey {
			return attr.Val, true
		}
	}

	return "", false
}
