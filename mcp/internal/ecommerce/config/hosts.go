package config

import "net/url"

type Host struct {
	Name  string
	Group string
}

const (
	AUDIO      = "audio"
	AUDIOPHILE = "audiophile"
	KEYBOARDS  = "keyboards"
	DIY        = "diy"
	AMAZON = "amazon.in"
)

func Electronics() []Host {
	return []Host{
		{"boat-lifestyle.com", AUDIO},
		{"gonoise.com", AUDIO},
		{"goboult.com", AUDIO},
		{"mivi.in", AUDIO},
		{"zebronics.com", AUDIO},
		{"portronics.com", AUDIO},
		{"ambraneindia.com", AUDIO},
		{"hammeronline.in", AUDIO},
		{"crossbeats.com", AUDIO},
		{"headphonezone.in", AUDIOPHILE},
		{"conceptkart.com", AUDIOPHILE},
		{"kreo-tech.com", KEYBOARDS},
		{"meckeys.com", KEYBOARDS},
		{"stackskb.com", KEYBOARDS},
		{"neomacro.in", KEYBOARDS},
		{"genesispc.in", KEYBOARDS},
		{"vishalperipherals.com", KEYBOARDS},
		{"shopatsc.com", DIY},
		{"robocraze.com", DIY},
		{"makerbazar.in", DIY},
		{"quartzcomponents.com", DIY},
	}
}

func AmazonSearchUrl(query string) string {
	return "https://www.amazzon.in/s?k=" + url.QueryEscape(query)
}
