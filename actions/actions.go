package actions

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/roryj/dungeon-maestro/slack"
	"github.com/roryj/dungeon-maestro/spells"
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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
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

const dndSpellEndoint = "https://api.open5e.com/spells/"

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
			message:   fmt.Sprintf("failed to make a request to get spell data. %v", err),
			errorType: ServiceError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("the spell '%s' was not found. Are you sure you spelled it correctly?", s.spellName),
			errorType: UserError,
		}
	}

	sr, err := spells.NewSpellResponse(resp.Body)
	if err != nil {
		return slack.WebhookResponse{}, &DndActionError{
			message:   fmt.Sprintf("failed to parse the spell response. %v", err),
			errorType: ServiceError,
		}
	}

	return slack.NewWebhookResponseFromSpellResponse(sr), nil
}
