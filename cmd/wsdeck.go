// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const wsdeckUrl = "https://wsdecks.com"

// Card contains card info
type Card struct {
	ID     string
	Amount int
	Level  int
	Color  string
}

// Deck is a deck
type Deck struct {
	Cards       []Card
	Name        string
	Description string
}

func getCardDeckInfo(url string) (Deck, error) {

	var deck = Deck{}
	var cardsDeck = []Card{}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return deck, err
	}
	deck.Description = doc.Find("blockquote").Text()
	deck.Name = doc.Find(".bcMain h3").Text()
	doc.Find("div .wscard").Each(func(i int, s *goquery.Selection) {
		card := Card{}
		cardID, exists := s.Attr("data-cardid")
		if exists {
			var split = strings.Split(cardID, "-")
			card.ID = fmt.Sprintf("%s%s%s", split[0], "-", strings.Replace(split[1], "E", "", 1))
		}

		cardAmount, exists := s.Attr("data-amount")
		if exists {
			card.Amount, err = strconv.Atoi(cardAmount)
			if err != nil {
				fmt.Println(err)
			}
		}

		level, exists := s.Attr("data-level")
		if exists {
			card.Level, err = strconv.Atoi(level)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			card.Level = 42
		}

		card.Color = s.AttrOr("data-color", "undefined")
		cardsDeck = append(cardsDeck, card)
	})
	deck.Cards = cardsDeck
	return deck, nil
}

// GetDecks from wsdeck url
func GetDecks(url string, max int) []Deck {
	var decks = []Deck{}
	for n := 1; n < max; n++ {
		pageURL := fmt.Sprintf("%v&page=%v", url, n)
		log.Printf("Get decks from %v", pageURL)
		doc, err := goquery.NewDocument(pageURL)
		if err != nil {
			log.Fatalf("Unable to parse wsdeck url %v", url)
		}
		doc.Find(".deckList .deckname a").Each(func(number int, deckA *goquery.Selection) {
			url := strings.Join([]string{wsdeckUrl, deckA.AttrOr("href", "undefined")}, "")
			deck, err := getCardDeckInfo(url)
			if err != nil {
				log.Fatalf("Error on parsing %v deck", url)
			}

			decks = append(decks, deck)
		})
	}

	return decks
}
