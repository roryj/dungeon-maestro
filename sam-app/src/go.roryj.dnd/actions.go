package main

import (
	"fmt"
	"golang.org/x/net/html"
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
	log.Printf("Processing dice roll on %v", d)
	var total int

	for i := 0; i < d.numberOfDice; i++ {
		total += rand.Intn(d.diceSides) + 1
	}

	return fmt.Sprintf("Rolled %d d%d and got %d\n", d.numberOfDice, d.diceSides, total), nil
}

const dndSpellEndoint = "https://www.dndbeyond.com/spells/"

type IdentifySpell struct {
	spellName  string
	statBlocks []string
}

func NewIdentifySpell(input string) (*IdentifySpell, error) {
	s := strings.Replace(input, "+", " ", -1)

	return &IdentifySpell{
		spellName: s,
		statBlocks: []string{},
	}, nil
}

// IdentifySpell#ProcessAction gets info about a spell from internet sources. All spaces and "/" in spell names should be
// replaced by "-". Any apostrophes should be removed completely
// ex: 	Antimagic Field -> /spells/antimagic-field
// 		Antipathy/Sympathy -> /spells/antipathy-sympathy
func (s *IdentifySpell) ProcessAction() (string, error) {
	log.Printf("Processing spell identify on %v", s)

	urlEncodedSpell := strings.Replace(s.spellName, " ", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "/", "-", -1)
	urlEncodedSpell = strings.Replace(urlEncodedSpell, "'", "", -1)
	urlEncodedSpell = strings.ToLower(urlEncodedSpell)

	fullPath := dndSpellEndoint + urlEncodedSpell

	log.Printf("url: %s", fullPath)

	resp, err := http.Get(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse the request from dnd beyod. %v", err)
	}
	defer resp.Body.Close()

	rootDoc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse the request from dnd beyod. %v", err)
	}

	var missingAttributes []string
	result := fmt.Sprintf("Description of %s\n", s.spellName)

	spellAttributes := []string{"Level", "Casting Time", "Range/Area", "Components", "Duration", "School",
		"Attack/Save", "Damage/Effect"}
	//spellAttributes := []string { "Components" }

	for _, a := range spellAttributes {
		log.Printf("getting %s attribute", a)
		r, ok := getSpellAttribute(rootDoc, a)
		if !ok {
			missingAttributes = append(missingAttributes, a)
		} else {
			result = result + r + "\n"
		}
	}

	if len(missingAttributes) > 0 {
		return "", fmt.Errorf("failed to find attributes for: %v", missingAttributes)
	}

	return result, nil
}

func getSpellAttribute(n *html.Node, spellAttribute string) (string, bool) {
	value, ok := getValueForAttribute(n, spellAttribute)
	return fmt.Sprintf("%s: %s", spellAttribute, value), ok
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
