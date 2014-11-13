package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
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
	values := []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	return values[int(v)]
}

type Card struct {
	Suit  CardSuit
	Value CardValue
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Suit, c.Value)
}

var NewShoe []Card = []Card{
	Card{Spades, Ace}, Card{Spades, Two}, Card{Spades, Three}, Card{Spades, Four}, Card{Spades, Five}, Card{Spades, Six}, Card{Spades, Seven}, Card{Spades, Eight}, Card{Spades, Nine}, Card{Spades, Ten}, Card{Spades, Jack}, Card{Spades, Queen}, Card{Spades, King},
	Card{Hearts, Ace}, Card{Hearts, Two}, Card{Hearts, Three}, Card{Hearts, Four}, Card{Hearts, Five}, Card{Hearts, Six}, Card{Hearts, Seven}, Card{Hearts, Eight}, Card{Hearts, Nine}, Card{Hearts, Ten}, Card{Hearts, Jack}, Card{Hearts, Queen}, Card{Hearts, King},
	Card{Diamonds, Ace}, Card{Diamonds, Two}, Card{Diamonds, Three}, Card{Diamonds, Four}, Card{Diamonds, Five}, Card{Diamonds, Six}, Card{Diamonds, Seven}, Card{Diamonds, Eight}, Card{Diamonds, Nine}, Card{Diamonds, Ten}, Card{Diamonds, Jack}, Card{Diamonds, Queen}, Card{Diamonds, King},
	Card{Clubs, Ace}, Card{Clubs, Two}, Card{Clubs, Three}, Card{Clubs, Four}, Card{Clubs, Five}, Card{Clubs, Six}, Card{Clubs, Seven}, Card{Clubs, Eight}, Card{Clubs, Nine}, Card{Clubs, Ten}, Card{Clubs, Jack}, Card{Clubs, Queen}, Card{Clubs, King},
}

type HandState int

const (
	Playing HandState = iota
	Standing
	TwentyOne
	Busted
)

type Hand struct {
	Cards []Card
	State HandState
}

func (h Hand) String() string {
	cardStrings := make([]string, len(h.Cards))
	for i, c := range h.Cards {
		cardStrings[i] = c.String()
	}
	return strings.Join(cardStrings, "  ")
}

func (h *Hand) Value() int {
	total := 0
	aces := 0
	for _, c := range h.Cards {
		// TODO: handle aces
		if c.Value == Ace {
			aces += 1
		} else {
			total += c.Value.Value()
		}
	}

	for 0 < aces && 21 < total + 11 * aces {
		total += 1
		aces -= 1
	}
	total += 11 * aces
	return total
}

var Game struct {
	Shoe []Card
	Hands map[string]Hand
	DealerHand Hand
	Dealing bool
}

func InitCards() {
	Game.Hands = make(map[string]Hand)
	Game.DealerHand = Hand{}
	Game.Dealing = true
}

func Deal() (card Card, reshuffled bool) {
	if len(Game.Shoe) == 0 {
		reshuffled = true
		Game.Shoe = NewShoe
	}

	i := rand.Intn(len(Game.Shoe))
	card = Game.Shoe[i]
	lastIndex := len(Game.Shoe) - 1
	Game.Shoe[i] = Game.Shoe[lastIndex]
	Game.Shoe = Game.Shoe[:lastIndex]

	return
}

func JackDeal(msg *SlackMessage) (string, error) {
	if !Game.Dealing {
		return "I can deal you in next round.", nil
	}
	if hand, ok := Game.Hands[msg.UserName]; ok {
		return fmt.Sprintf("You're already dealt in this round.\n\nYour hand:  %s", hand), nil
	}

	card1, shuf1 := Deal()
	card2, shuf2 := Deal()
	reshuffled := shuf1 || shuf2

	hand := Hand{}
	hand.Cards = []Card{card1, card2}
	text := fmt.Sprintf("Your hand:  %s  %s", card1, card2)

	if hand.Value() == 21 {
		hand.State = TwentyOne
		text = fmt.Sprintf("%s\n\nBlackjack!", text)
	}

	Game.Hands[msg.UserName] = hand

	if len(Game.DealerHand.Cards) == 0 {
		card1, shuf1 = Deal()
		card2, shuf2 = Deal()
		reshuffled = reshuffled || shuf1 || shuf2

		Game.DealerHand = Hand{}
		Game.DealerHand.Cards = []Card{card1, card2}
		text = fmt.Sprintf("%s\n\nDealer's hand:  🂠  %s", text, card2)
	}

	if reshuffled {
		text = fmt.Sprintf("Let me reshuffle the shoe...\n\n%s", text)
	}

	return text, nil
}

func JackWinner(msg *SlackMessage, text string) (string, error) {
	// Who had 21s?
	var topStanding []string
	var topStandingValue int

	log.Println("Who won among", len(Game.Hands), "& dealer?")
	for userId, hand := range Game.Hands {
		if hand.State == Busted {
			continue
		}

		if topStandingValue < hand.Value() {
			topStanding = []string{userId}
			topStandingValue = hand.Value()
		} else if topStandingValue == hand.Value() {
			topStanding = append(topStanding, userId)
		}
	}
	log.Println("Found live winners", topStanding, "with score", topStandingValue)
	log.Println("Dealer is in state", Game.DealerHand.State, "with value", Game.DealerHand.Value())

	if Game.DealerHand.State != Busted && topStandingValue < Game.DealerHand.Value() {
		text = fmt.Sprintf("%s\n\nDealer won this round.", text)
		log.Println("So dealer won!")
	} else {
		switch len(topStanding) {
		case 0:
			log.Println("So no one won!")
			text = fmt.Sprintf("%s\n\nNo winner this round.", text)
		case 1:
			log.Println("So", topStanding[0], "won by themself!")
			text = fmt.Sprintf("%s\n\n%s won this round.", text, topStanding[0])
		default:
			log.Println("So", topStanding, "all won!")

			lastIndex := len(topStanding)-1
			lastWinner := topStanding[lastIndex]
			topStanding = topStanding[:lastIndex]
			winners := strings.Join(topStanding, ", ")

			text = fmt.Sprintf("%s\n\n%s & %s won this round.", text, winners, lastWinner)
		}
	}

	InitCards()

	return text, nil
}

func JackContinue(msg *SlackMessage, text string) (string, error) {
	for _, h := range Game.Hands {
		if h.State == Playing {
			return text, nil
		}
	}

	// Nobody is in Playing state, so the dealer can finish playing.
	reshuffled := false
	dealer := Game.DealerHand
	for dealer.State == Playing {
		card, shuf := Deal()
		reshuffled = reshuffled || shuf
		dealer.Cards = append(dealer.Cards, card)

		switch x := dealer.Value(); {
		case 21 < x:
			dealer.State = Busted
		case 21 == x:
			dealer.State = TwentyOne
		case 17 <= x:
			dealer.State = Standing
		}
	}
	Game.DealerHand = dealer

	text = fmt.Sprintf("%s\n\nDealer's hand:  %s", text, dealer)
	switch dealer.State {
	case Busted:
		text = fmt.Sprintf("%s\n\nBusted!", text)
	case TwentyOne:
		text = fmt.Sprintf("%s\n\nTwenty-one!", text)
	}

	return JackWinner(msg, text)
}

func JackStand(msg *SlackMessage) (string, error) {
	userId := msg.UserName
	hand, ok := Game.Hands[userId]
	if !ok {
		if Game.Dealing {
			return "I didn't deal you in, use `deal` to join.", nil
		}
		return "I can deal you in next round.", nil
	}

	// I was already dealt, but I asked to stand, so the deal phase is over.
	if !Game.Dealing {
		Game.Dealing = false
	}

	if hand.State != Playing {
		return "You can't stand now.", nil
	}

	hand.State = Standing
	Game.Hands[userId] = hand

	text := fmt.Sprintf("Sure, standing at %d.", hand.Value())

	return JackContinue(msg, text)
}

func JackHit(msg *SlackMessage) (string, error) {
	userId := msg.UserName
	hand, ok := Game.Hands[userId]
	if !ok {
		// Not dealt yet, let's go ahead and deal.
		text, err := JackDeal(msg)
		if err != nil {
			return text, err
		}
		if Game.Dealing {
			text = fmt.Sprintf("Okay, I'll deal you in.\n\n%s", text)
		}
		return text, err
	}

	// I was already dealt, but I asked to be hit, so the deal phase is over.
	if !Game.Dealing {
		Game.Dealing = false
	}

	card, reshuffled := Deal()

	hand.Cards = append(hand.Cards, card)

	total := hand.Value()
	text := fmt.Sprintf("Your hand:  %s", hand)

	if total < 21 {
	} else if total == 21 {
		text = fmt.Sprintf("%s\n\nTwenty-one!", text)
		hand.State = TwentyOne
	} else {
		text = fmt.Sprintf("%s\n\nBusted!", text)
		hand.State = Busted
	}

	Game.Hands[userId] = hand

	if reshuffled {
		text = fmt.Sprintf("Let me reshuffle the shoe...\n\n%s", text)
	}

	return JackContinue(msg, text)
}
