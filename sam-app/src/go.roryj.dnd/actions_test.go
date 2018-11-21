package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewIdentifySpell(t *testing.T) {
	spell := IdentifySpell{
		spellName:  "Abi Dalzims Horrid Wilting",
		statBlocks: []string{},
	}

	result, err := spell.ProcessAction()

	assert.Nil(t, err, "the result should not throw an error")
	t.Logf("Result: %s. Error: %v", result, err)
}
