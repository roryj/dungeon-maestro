package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

type DndAction interface {
	ProcessAction() (string, error)
}

type DiceRoll struct {
	diceSides    int
	numberOfDice int
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
		diceSides:    diceSides,
		numberOfDice: numDice,
	}, nil
}

func (d *DiceRoll) ProcessAction() (string, error) {

	var total int

	for i := 0; i < d.numberOfDice; i++ {
		total += rand.Intn(d.diceSides) + 1
	}

	return fmt.Sprintf("Rolled %d d%d and got %d\n", d.numberOfDice, d.diceSides, total), nil
}

const dndSpellEndoint = "https://www.dndbeyond.com/spells/"

type IdentifySpell struct {
	spellName string
	statBlocks []string
}

func NewIdentifySpell(input string) (*IdentifySpell, error) {
	// split on html encoded spaces (+)
	split := strings.Split(input, "+")

	if len(split) < 1 {
		return &IdentifySpell{}, fmt.Errorf("incorrect number of arguments for identify. You need put a spell to identify")
	}

	var statBlocks []string

	if len(split) > 1 {
		statBlocks = split[1:]
	}

	return &IdentifySpell{
		spellName: split[0],
		statBlocks: statBlocks,
	}, nil
}

// IdentifySpell#ProcessAction gets info about a spell from internet sources. All spaces and "/" in spell names should be
// replaced by "-". Any apostrophes should be removed completely
// ex: 	Antimagic Field -> /spells/antimagic-field
// 		Antipathy/Sympathy -> /spells/antipathy-sympathy
func (s *IdentifySpell) ProcessAction() (string, error) {

	urlEncodedSpell := strings.Replace(s.spellName, " ", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "/", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "'", "", -1)

	fullPath := dndSpellEndoint + urlEncodedSpell

	resp, err := http.Get(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse the request from dnd beyod. %v", err)
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	rootDoc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse the request from dnd beyod. %v", err)
	}

	log.Printf("Result: %v", rootDoc)

	return fmt.Sprintf("%s", string(res)), nil
}


