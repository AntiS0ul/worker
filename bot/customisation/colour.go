package customisation

import (
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/dbclient"
)

type Colour int16

func (c Colour) Int16() int16 {
	return int16(c)
}

const (
	Green Colour = iota
	Red
	Orange
	Lime
	Blue
)

var defaultColours = map[Colour]int{
	Green:  0x2ECC71,
	Red:    0xFC3F35,
	Orange: 16740864,
	Lime:   7658240,
	Blue:   472219,
}

func GetDefaultColour(colour Colour) int {
	return defaultColours[colour]
}

func GetColours(guildId uint64) (map[Colour]int, error) {
	raw, err := dbclient.Client.CustomColours.GetAll(guildId)
	if err != nil {
		return defaultColours, err
	}

	colours := make(map[Colour]int)
	for id, hex := range raw {
        colours[Colour(id)] = hex
    }

	for id, hex := range defaultColours {
        if _, ok := colours[id]; !ok {
            colours[id] = hex
        }
    }

	return colours, nil
}

func GetColour(guildId uint64, colourCode Colour) (int, error) {
	colour, ok, err := dbclient.Client.CustomColours.Get(guildId, colourCode.Int16())
	if err != nil {
        return 0, err
    }

	if !ok {
		return GetDefaultColour(colourCode), nil
	}

	return colour, nil
}


func GetColourOrDefault(guildId uint64, colourCode Colour) int {
	colour, ok, err := dbclient.Client.CustomColours.Get(guildId, colourCode.Int16())
	if err != nil {
		sentry.Error(err)
        return GetDefaultColour(colourCode)
    }

	if !ok {
		return GetDefaultColour(colourCode)
	}

	return colour
}