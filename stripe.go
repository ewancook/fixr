package fixr

import (
	"bytes"
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
	Error *stripeError `json:"error,omitempty"`
}

func (t *token) error() error {
	if t.Error != nil {
		return fmt.Errorf("%s (code: %s; type: %s; param: %s)", t.Error.Message, t.Error.Code, t.Error.Type, t.Error.Param)
	}
	return nil
}

func (t *token) clearError() {
	t.Error = nil
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
	apiError
	User *stripeUser `json:"stripe_user"`
}

// HasCard checks for the existence of a saved card in the user's FIXR account.
// The result and an error, if encountered, will be returned.
func (c *Client) HasCard() (existingCards bool, returnErr error) {
	if c.StripeUser == nil {
		return
	}
	if err := c.get(meURL, true, c); err != nil {
		returnErr = errors.Wrap(err, "error updating stripe details")
	}
	existingCards = len(c.StripeUser.Cards) != 0
	return
}

func encodeAddCardPayload(pl payload) (*bytes.Buffer, error) {
	data := new(bytes.Buffer)
	urlValues, err := buildURLValues(pl)
	if err != nil {
		return nil, err
	}
	data.WriteString(urlValues.Encode())
	return data, nil
}

// AddCard saves a card to the user's FIXR account, given the card details.
// An error will be returned if encountered
func (c *Client) AddCard(num, month, year, cvc, zip string) error {
	token := token{}
	pl := payload{
		"payment_user_agent": ua,
		"key":                key,
		"card[number]":       num,
		"card[exp_month]":    month,
		"card[exp_year]":     year,
		"card[cvc]":          cvc,
		"card[address_zip]":  zip,
	}
	data, err := encodeAddCardPayload(pl)
	if err != nil {
		return err
	}
	if err := c.post(cardURL, data, false, &token); err != nil {
		return errors.Wrap(err, "error retrieving tokens")
	}
	tokenReq := tokenRequest{}
	tokenData, err := jsonifyPayload(payload{"token": &token.Token})
	if err != nil {
		return err
	}
	if err := c.post(tokenURL, tokenData, true, &tokenReq); err != nil {
		return errors.Wrap(err, "error sending tokens")
	}
	c.StripeUser = tokenReq.User
	return nil
}
