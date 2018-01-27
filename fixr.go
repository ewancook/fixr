package fixr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const (
	bookingURL = "https://api.fixr-app.com/api/v2/app/booking"
	promoURL   = "https://api.fixr-app.com/api/v2/app/promo_code/%d/%s"
	loginURL   = "https://api.fixr-app.com/api/v2/app/user/authenticate/with-email"
	eventURL   = "https://api.fixr-app.com/api/v2/app/event/%d"
	cardURL    = "https://api.stripe.com/v1/tokens"
	tokenURL   = "https://api.fixr-app.com/api/v2/app/stripe"
)

type payload map[string]interface{}

type client struct {
	Email      string
	Password   string
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MagicURL   string `json:"magic_login_url"`
	AuthToken  string `json:"auth_token"`
	httpClient *http.Client
	Error      string `json:"message,omitempty"`
}

type event struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Tickets []ticket `json:"tickets"`
	Error   string   `json:"detail,omitempty"`
}

type ticket struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Type       int     `json:"type"`
	Currency   string  `json:"currency"`
	Price      float64 `json:"price"`
	BookingFee float64 `json:"booking_fee"`
	Max        int     `json:"max_per_user"`
	SoldOut    bool    `json:"sold_out"`
	Expired    bool    `json:"expired"`
	Invalid    bool    `json:"not_yet_valid"`
}

type promoCode struct {
	Code       string  `json:"code"`
	Price      float64 `json:"price"`
	BookingFee float64 `json:"booking_fee"`
	Currency   string  `json:"currency"`
	Max        int     `json:"max_per_user"`
	Remaining  int     `json:"remaining"`
	Error      string  `json:"message,omitempty"`
}

type booking struct {
	Event event  `json:"event"`
	Name  string `json:"user_full_name"`
	PDF   string `json:"pdf"`
	State int    `json:"state"`
	Error string `json:"message,omitempty"`
}

func NewClient(email, pass string) *client {
	return &client{Email: email, Password: pass, httpClient: new(http.Client)}
}

func (c *client) get(addr string, auth bool, obj interface{}) error {
	r, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return errors.New("error creating GET request")
	}
	return c.req(r, auth, obj)
}

func (c *client) post(addr string, val payload, auth bool, obj interface{}) error {
	data := new(bytes.Buffer)
	if addr == cardURL {
		pl := url.Values{}
		for x, y := range val {
			str, ok := y.(string)
			if !ok {
				return errors.New("failed to build payload")
			}
			pl.Set(x, str)
		}
		data.WriteString(pl.Encode())
	} else {
		if err := json.NewEncoder(data).Encode(val); err != nil {
			return errors.Wrap(err, "error jsonifying payload")
		}
	}
	r, err := http.NewRequest("POST", addr, data)
	if err != nil {
		return errors.New("error creating POST request")
	}
	return c.req(r, auth, obj)
}

func (c *client) req(r *http.Request, auth bool, obj interface{}) error {
	if auth {
		r.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthToken))
	}
	if r.URL.String() == cardURL {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		r.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return errors.Wrap(err, "error executing request")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

func (c *client) Logon() error {
	pl := payload{
		"email":    c.Email,
		"password": c.Password,
	}
	if err := c.post(loginURL, pl, false, c); err != nil {
		return errors.Wrap(err, "error logging on")
	}
	if c.Error != "" {
		return errors.New(fmt.Sprintf("error logging on: %s", c.Error))
	}
	return nil
}

func (c *client) Event(n int) (*event, error) {
	e := new(event)
	if err := c.get(fmt.Sprintf(eventURL, n), false, e); err != nil {
		return nil, errors.Wrap(err, "error getting event")
	}
	if e.Error != "" {
		return nil, errors.New(fmt.Sprintf("error getting event: %s", e.Error))
	}
	return e, nil
}

func (c *client) Promo(ticket_id int, s string) (*promoCode, error) {
	p := new(promoCode)
	if err := c.get(fmt.Sprintf(promoURL, ticket_id, s), true, p); err != nil {
		return nil, errors.Wrap(err, "error getting promo code")
	}
	if p.Error != "" {
		return nil, errors.New(fmt.Sprintf("error getting promo code: %s", p.Error))
	}
	return p, nil
}

func (c *client) Book(ticket_id, amount int, promo *promoCode, gen bool) (*booking, error) {
	b := new(booking)
	pl := payload{
		"ticket_id": ticket_id,
		"amount":    amount,
	}
	if gen {
		pl["purchase_key"] = genKey()
	}
	if promo != nil {
		pl["promo_code"] = promo.Code
	}
	if err := c.post(bookingURL, pl, true, b); err != nil {
		return nil, errors.Wrap(err, "error booking ticket")
	}
	if b.Error != "" {
		return nil, errors.New(fmt.Sprintf("error booking ticket: %s", b.Error))
	}
	return b, nil
}
