package main

import (
	"fmt"
	"math/rand"
)

type CardSuit int
type CardValue int

const (
	Spades CardSuit = iota
	Hearts
	Diamonds
	Clubs
	numSuits int = iota
)

func (s CardSuit) String() string {
	suits := []string{":spades:", ":hearts:", ":diamonds:", ":clubs:"}
	return suits[int(s)]
}

const (
	_ CardValue = iota
	Ace
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	numValues int = iota
)

func (v CardValue) Value() int {
	if v < Jack {
		return int(v)
	}
	return 10
}

func (v CardValue) String() string {
	values := []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "J", "Q", "K"}
	return values[int(v)]
}

type Card struct {
	Suit  CardSuit
	Value CardValue
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Suit, c.Value)
}

var Shoe []Card

var NewShoe []Card = []Card{
	Card{Spades, Ace}, Card{Spades, Two}, Card{Spades, Three}, Card{Spades, Four}, Card{Spades, Five}, Card{Spades, Six}, Card{Spades, Seven}, Card{Spades, Eight}, Card{Spades, Nine}, Card{Spades, Ten}, Card{Spades, Jack}, Card{Spades, Queen}, Card{Spades, King},
	Card{Hearts, Ace}, Card{Hearts, Two}, Card{Hearts, Three}, Card{Hearts, Four}, Card{Hearts, Five}, Card{Hearts, Six}, Card{Hearts, Seven}, Card{Hearts, Eight}, Card{Hearts, Nine}, Card{Hearts, Ten}, Card{Hearts, Jack}, Card{Hearts, Queen}, Card{Hearts, King},
	Card{Diamonds, Ace}, Card{Diamonds, Two}, Card{Diamonds, Three}, Card{Diamonds, Four}, Card{Diamonds, Five}, Card{Diamonds, Six}, Card{Diamonds, Seven}, Card{Diamonds, Eight}, Card{Diamonds, Nine}, Card{Diamonds, Ten}, Card{Diamonds, Jack}, Card{Diamonds, Queen}, Card{Diamonds, King},
	Card{Clubs, Ace}, Card{Clubs, Two}, Card{Clubs, Three}, Card{Clubs, Four}, Card{Clubs, Five}, Card{Clubs, Six}, Card{Clubs, Seven}, Card{Clubs, Eight}, Card{Clubs, Nine}, Card{Clubs, Ten}, Card{Clubs, Jack}, Card{Clubs, Queen}, Card{Clubs, King},
}

var Hands map[string][]Card

func InitCards() {
	Hands = make(map[string][]Card)
}

func Deal() (card Card, reshuffled bool) {
	if len(Shoe) == 0 {
		reshuffled = true
		Shoe = NewShoe
	}

	i := rand.Intn(len(Shoe))
	card = Shoe[i]
	lastIndex := len(Shoe) - 1
	Shoe[i] = Shoe[lastIndex]
	Shoe = Shoe[:len(Shoe)-1]

	return
}

func JackDeal(msg *SlackMessage) (string, error) {
	card1, shuf1 := Deal()
	card2, shuf2 := Deal()
	reshuffled := shuf1 || shuf2

	Hands[msg.UserId] = []Card{card1, card2}

	text := fmt.Sprintf("Your hand:  %s  %s", card1, card2)
	// TODO: handle blackjacks with Aces (which never count as 11 yet)

	if reshuffled {
		text = fmt.Sprintf("Let me reshuffle the shoe...\n\n%s", text)
	}

	return text, nil
}

func JackHit(msg *SlackMessage) (string, error) {
	userId := msg.UserId
	hand, ok := Hands[userId]
	if !ok {
		// Not dealt yet, let's go ahead and deal.
		text, err := JackDeal(msg)
		if err != nil {
			return text, err
		}
		return fmt.Sprintf("Okay, I'll deal you in.\n\n%s", text), err
	}

	card, reshuffled := Deal()

	hand = append(hand, card)
	total := 0
	text := "Your hand:"
	for _, c := range hand {
		// TODO: handle aces
		total += c.Value.Value()
		text = fmt.Sprintf("%s  %s", text, c)
	}

	if total < 21 {
		Hands[userId] = hand
	} else if total == 21 {
		delete(Hands, userId)
		text = fmt.Sprintf("%s\n\nBlackjack!", text)
	} else {
		delete(Hands, userId)
		text = fmt.Sprintf("%s\n\nBusted!", text)
	}

	if reshuffled {
		text = fmt.Sprintf("Let me reshuffle the shoe...\n\n%s", text)
	}

	return text, nil
}
