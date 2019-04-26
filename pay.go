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

// HasCard checks for the existence of a saved card in the user's FIXR account.
// The result and an error, if encountered, will be returned.
func (c *Client) HasCard() (bool, error) {
	if c.StripeUser == nil {
		return false, nil
	}
	existing := len(c.StripeUser.Cards) != 0
	if err := c.get(meURL, true, c); err != nil {
		return existing, errors.Wrap(err, "error updating stripe details")
	}
	if len(c.Error) > 0 {
		return existing, fmt.Errorf("error updating stripe details: %s", c.Error)
	}
	return len(c.StripeUser.Cards) != 0, nil
}

// AddCard saves a card to the user's FIXR account, given the card details.
// An error will be returned if encountered
func (c *Client) AddCard(num, month, year, cvc, zip string) error {
	token := new(token)
	pl := payload{
		"payment_user_agent": ua,
		"key":                key,
		"card[number]":       num,
		"card[exp_month]":    month,
		"card[exp_year]":     year,
		"card[cvc]":          cvc,
		"card[address_zip]":  zip,
	}
	if err := c.post(cardURL, pl, false, token); err != nil {
		return errors.Wrap(err, "error retrieving tokens")
	}
	if token.Error != (stripeError{}) {
		return fmt.Errorf("error retrieving tokens: %s", token.Error.Message)
	}
	tokenReq := new(tokenRequest)
	if err := c.post(tokenURL, payload{"token": token.Token}, true, tokenReq); err != nil {
		return errors.Wrap(err, "error sending tokens")
	}
	if len(tokenReq.Error) > 0 {
		return fmt.Errorf("error sending tokens: %s", tokenReq.Error)
	}
	c.StripeUser = &tokenReq.User
	return nil
}
