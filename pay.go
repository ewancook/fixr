package fixr

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	key string = "pk_live_Jc9zYhZyq3a4JviHWZFBFRdp"
	ua  string = "stripe.js/604c5e8"
)

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

type tokenRequest struct {
	User  stripeUser `json:"stripe_user"`
	Error string     `json:"message,omitempty"`
}

func (c *client) HasCard() (bool, error) {
	if c.StripeUser == nil {
		return false, nil
	}
	existing := len(c.StripeUser.Cards) != 0
	if err := c.get(meURL, true, c); err != nil {
		return existing, errors.Wrap(err, "error updating stripe details")
	}
	if c.Error != "" {
		return existing, fmt.Errorf("error updating stripe details: %s", c.Error)
	}
	return len(c.StripeUser.Cards) != 0, nil
}

func (c *client) AddCard(num, month, year, cvc, zip string) (*stripeUser, error) {
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
	if err := c.post(cardURL, pl, false, t); err != nil {
		return nil, errors.Wrap(err, "error retrieving tokens")
	}
	if t.Error != (stripeError{}) {
		m := fmt.Errorf("error retrieving tokens: %s", t.Error.Message)
		return nil, m
	}
	s := new(tokenRequest)
	if err := c.post(tokenURL, payload{"token": t.Token}, true, s); err != nil {
		return nil, errors.Wrap(err, "error sending tokens")
	}
	if s.Error != "" {
		return nil, fmt.Errorf("error sending tokens: %s", s.Error)
	}
	return &s.User, nil
}
