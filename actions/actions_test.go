package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIdentifySpell(t *testing.T) {
	spell := IdentifySpell{
		spellName:  "Abi Dalzims Horrid Wilting",
		statBlocks: []string{},
	}

	result, err := spell.ProcessAction()

	assert.Nil(t, err, "the result should not throw an error")
	t.Logf("Result: %v. Error: %v", result, err)
}
