package fixr

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/pkg/errors"
)

const (
	key string = "pk_live_Jc9zYhZyq3a4JviHWZFBFRdp"
	ua  string = "stripe.js/604c5e8"
)

type card struct {
	number, month, year, cvc, zip string
}

type stripeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Param   string `json:"param"`
	Code    string `json:"code,omitempty"`
}

type token struct {
	Token string `json:"id"`
	Card  struct {
		ID string `json:"id"`
	} `json:"card"`
	Error stripeError `json:"error,omitempty"`
}

type stripeUser struct {
	UserID string       `json:"stripe_id"`
	Cards  []stripeCard `json:"cards"`
}

type stripeCard struct {
	CardID      string `json:"stripe_id"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpiryMonth int    `json:"exp_month"`
	ExpiryYear  int    `json:"exp_year"`
	Country     string `json:"country"`
}

func (c *client) NewCard(num, month, year, cvc, zip string) (*token, error) {
	t := new(token)
	pl := payload{
		"payment_user_agent": ua,
		"key":                key,
		"card[number]":       num,
		"card[exp_month]":    month,
		"card[exp_year]":     year,
		"card[cvc]":          cvc,
		"card[address_zip]":  zip,
	}
	if err := c.req("POST", cardURL, pl, false, t); err != nil {
		return nil, errors.Wrap(err, "error retrieving tokens")
	}
	if t.Error != (stripeError{}) {
		m := fmt.Sprintf("error retrieving tokens: %s", t.Error.Message)
		return nil, errors.New(m)
	}
	return t, nil
}

func (c *client) Tokens(t string) (*stripeUser, error) {
	var s struct {
		User  stripeUser `json:"stripe_user"`
		Error string     `json:"message,omitempty"`
	}
	pl := payload{"token": t}
	if err := c.req("POST", tokenURL, pl, true, &s); err != nil {
		return nil, errors.Wrap(err, "error sending tokens")
	}
	if s.Error != "" {
		return nil, errors.New(fmt.Sprintf("error sending tokens: %s", s.Error))
	}
	return &s.User, nil
}

func uuid() string {
	a := math.Floor(65536 * (1 + rand.Float64()))
	return strconv.FormatInt(int64(a), 16)[1:]
}

func genKey() string {
	s := make([]interface{}, 8)
	for x := 0; x < 8; x++ {
		s[x] = uuid()
	}
	return fmt.Sprintf("%s%s-%s-%s-%s-%s%s%s", s...)
}
