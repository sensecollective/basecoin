package types

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

func (coin Coin) String() string {
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

//regex codes for extracting coins from string
var reDenom = regexp.MustCompile("")
var reAmt = regexp.MustCompile("(\\d+)")

var reCoin = regexp.MustCompile("^([[:digit:]]+)[[:space:]]*([[:alpha:]]+)$")

func ParseCoin(str string) (Coin, error) {
	var coin Coin

	matches := reCoin.FindStringSubmatch(strings.TrimSpace(str))
	if matches == nil {
		return coin, errors.Errorf("%s is invalid coin definition", str)
	}

	// parse the amount (should always parse properly)
	amt, err := strconv.Atoi(matches[1])
	if err != nil {
		return coin, err
	}

	coin = Coin{matches[2], int64(amt)}
	return coin, nil
}

//----------------------------------------

type Coins []Coin

func (coins Coins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin.String())
	}
	return out[:len(out)-1]
}

func ParseCoins(str string) (Coins, error) {
	// empty string is empty list...
	if len(str) == 0 {
		return nil, nil
	}

	split := strings.Split(str, ",")
	var coins Coins

	for _, el := range split {
		coin, err := ParseCoin(el)
		if err != nil {
			return coins, err
		}
		coins = append(coins, coin)
	}

	// ensure they are in proper order, to avoid random failures later
	coins.Sort()
	if !coins.IsValid() {
		return nil, errors.Errorf("ParseCoins invalid: %#v", coins)
	}

	return coins, nil
}

// Must be sorted, and not have 0 amounts
func (coins Coins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true
	case 1:
		return coins[0].Amount != 0
	default:
		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if coin.Denom <= lowDenom {
				return false
			}
			if coin.Amount == 0 {
				return false
			}
			// we compare each coin against the last denom
			lowDenom = coin.Denom
		}
		return true
	}
}

// TODO: handle empty coins!
// Currently appends an empty coin ...
func (coinsA Coins) Plus(coinsB Coins) Coins {
	sum := []Coin{}
	indexA, indexB := 0, 0
	lenA, lenB := len(coinsA), len(coinsB)
	for {
		if indexA == lenA {
			if indexB == lenB {
				return sum
			} else {
				return append(sum, coinsB[indexB:]...)
			}
		} else if indexB == lenB {
			return append(sum, coinsA[indexA:]...)
		}
		coinA, coinB := coinsA[indexA], coinsB[indexB]
		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1:
			sum = append(sum, coinA)
			indexA += 1
		case 0:
			if coinA.Amount+coinB.Amount == 0 {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, Coin{
					Denom:  coinA.Denom,
					Amount: coinA.Amount + coinB.Amount,
				})
			}
			indexA += 1
			indexB += 1
		case 1:
			sum = append(sum, coinB)
			indexB += 1
		}
	}
	return sum
}

func (coins Coins) Negative() Coins {
	res := make([]Coin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, Coin{
			Denom:  coin.Denom,
			Amount: -coin.Amount,
		})
	}
	return res
}

func (coinsA Coins) Minus(coinsB Coins) Coins {
	return coinsA.Plus(coinsB.Negative())
}

func (coinsA Coins) IsGTE(coinsB Coins) bool {
	diff := coinsA.Minus(coinsB)
	if len(diff) == 0 {
		return true
	}
	return diff.IsNonnegative()
}

func (coins Coins) IsZero() bool {
	return len(coins) == 0
}

func (coinsA Coins) IsEqual(coinsB Coins) bool {
	if len(coinsA) != len(coinsB) {
		return false
	}
	for i := 0; i < len(coinsA); i++ {
		if coinsA[i] != coinsB[i] {
			return false
		}
	}
	return true
}

func (coins Coins) IsPositive() bool {
	if len(coins) == 0 {
		return false
	}
	for _, coinAmount := range coins {
		if coinAmount.Amount <= 0 {
			return false
		}
	}
	return true
}

func (coins Coins) IsNonnegative() bool {
	if len(coins) == 0 {
		return true
	}
	for _, coinAmount := range coins {
		if coinAmount.Amount < 0 {
			return false
		}
	}
	return true
}

/*** Implement Sort interface ***/

func (c Coins) Len() int           { return len(c) }
func (c Coins) Less(i, j int) bool { return c[i].Denom < c[j].Denom }
func (c Coins) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Coins) Sort()              { sort.Sort(c) }
