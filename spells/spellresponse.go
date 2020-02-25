package spells

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type SpellResponse struct {
	Name          string `json:"name"`
	Description   string `json:"desc"`
	HigherLevel   string `json:"higher_level"`
	Page          string `json:"page"`
	Range         string `json:"range"`
	Components    string `json:"components"`
	Material      string `json:"material"`
	Ritual        string `json:"ritual"`
	Duration      string `json:"duration"`
	Concentration string `json:"concentration"`
	CastingTime   string `json:"casting_time"`
	Level         string `json:"level"`
	School        string `json:"school"`
	DnDClass      string `json:"dnd_class"`
	Archetype     string `json:"archetype"`
	Circles       string `json:"circles"`
}

func NewSpellResponse(r io.Reader) (SpellResponse, error) {
	var sr SpellResponse

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return sr, err
	}

	err = json.Unmarshal(b, &sr)
	return sr, err
}
