package fixr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

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

var (
	fixrVersion     = "1.16.4"
	fixrPlatformVer = "Chrome/51.0.2704.103"
	userAgent       = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
)

type payload map[string]interface{}

type client struct {
	Email      string
	Password   string
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MagicURL   string `json:"magic_login_url"`
	AuthToken  string `json:"auth_token"`
	Error      string `json:"message,omitempty"`
	httpClient *http.Client
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

func Setup(seed bool) error {
	if seed {
		rand.Seed(time.Now().UnixNano())
	}
	if err := updateVersion(); err != nil {
		return errors.Wrap(err, "setup failed to update version")
	}
	return nil
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
	r.Header.Set("User-Agent", userAgent)
	if auth {
		r.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthToken))
	}
	if r.URL.String() == cardURL {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		r.Header.Set("Content-Type", "application/json")
		r.Header["FIXR-Platform"] = []string{"web"} // Circumvents canonical formatting
		r.Header["FIXR-Platform-Version"] = []string{fixrPlatformVer}
		r.Header["FIXR-App-Version"] = []string{fixrVersion}
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
		return fmt.Errorf("error logging on: %s", c.Error)
	}
	return nil
}

func (c *client) Event(n int) (*event, error) {
	e := new(event)
	if err := c.get(fmt.Sprintf(eventURL, n), false, e); err != nil {
		return nil, errors.Wrap(err, "error getting event")
	}
	if e.Error != "" {
		return nil, fmt.Errorf("error getting event: %s", e.Error)
	}
	return e, nil
}

func (c *client) Promo(ticketID int, s string) (*promoCode, error) {
	p := new(promoCode)
	if err := c.get(fmt.Sprintf(promoURL, ticketID, s), true, p); err != nil {
		return nil, errors.Wrap(err, "error getting promo code")
	}
	if p.Error != "" {
		return nil, fmt.Errorf("error getting promo code: %s", p.Error)
	}
	return p, nil
}

func (c *client) Book(ticket *ticket, amount int, promo *promoCode) (*booking, error) {
	b := new(booking)
	pl := payload{
		"ticket_id": ticket.ID,
		"amount":    amount,
	}
	for t, msg := range map[bool]string{
		ticket.SoldOut: "ticket selection has sold out",
		ticket.Expired: "ticket selection has expired",
		ticket.Invalid: "ticket selection is invalid",
	} {
		if t {
			return nil, errors.New(msg)
		}
	}
	if amount > ticket.Max {
		return nil, fmt.Errorf("cannot purchase more than the maximum (%d)", ticket.Max)
	}
	if ticket.BookingFee+ticket.Price > 0 {
		pl["purchase_key"] = genKey()
	}
	if promo != nil {
		pl["promo_code"] = promo.Code
	}
	if err := c.post(bookingURL, pl, true, b); err != nil {
		return nil, errors.Wrap(err, "error booking ticket")
	}
	if b.Error != "" {
		return nil, fmt.Errorf("error booking ticket: %s", b.Error)
	}
	return b, nil
}
