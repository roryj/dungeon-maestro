package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/roryj/dungeon-maestro/actions"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "roll",
				Usage: "roll the dice!",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "num-dice, n",
						Usage: "the number of dice to roll",
					},
					&cli.StringFlag{
						Name:  "num-sides, s",
						Usage: "the number of sides on the dice",
					},
				},
				Action: func(c *cli.Context) error {
					numDice := c.Int("num-dice")
					if numDice <= 0 {
						return fmt.Errorf("invalid num dice: %d", numDice)
					}

					numSides := c.Int("num-sides")
					if numSides <= 0 {
						return fmt.Errorf("invalid num sides: %d", numSides)
					}

					diceRoll, err := actions.NewDiceRoll("cli", fmt.Sprintf("%d+d%d", numDice, numSides))
					if err != nil {
						return err
					}

					r, err := diceRoll.ProcessAction()
					if err != nil {
						return err
					}

					fmt.Printf("%+v\n", r)
					return nil
				},
			},
			{
				Name:  "spell",
				Usage: "get information about a spell!",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "spell-name, s",
						Usage: "the spell name",
					},
				},
				Action: func(c *cli.Context) error {
					spell := c.String("spell-name")
					if spell == "" {
						return fmt.Errorf("invalid spell-name: %s", spell)
					}

					speller, err := actions.NewIdentifySpell(spell)
					if err != nil {
						return err
					}

					r, err := speller.ProcessAction()
					if err != nil {
						return err
					}

					fmt.Printf("%+v\n", r)
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
