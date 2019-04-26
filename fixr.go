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
	homeURL    = "https://fixr.co"
	bookingURL = "https://api.fixr-app.com/api/v2/app/booking"
	promoURL   = "https://api.fixr-app.com/api/v2/app/promo_code/%d/%s"
	loginURL   = "https://api.fixr-app.com/api/v2/app/user/authenticate/with-email"
	eventURL   = "https://api.fixr-app.com/api/v2/app/event/%d"
	cardURL    = "https://api.stripe.com/v1/tokens"
	tokenURL   = "https://api.fixr-app.com/api/v2/app/stripe"
	meURL      = "https://api.fixr-app.com/api/v2/app/user/me"
)

var (
	// FixrVersion represents FIXR's web API version.
	FixrVersion = "1.34.0"

	// FixrPlatformVer represents the "platform" used to "browse" the site.
	FixrPlatformVer = "Chrome/51.0.2704.103"

	// UserAgent is the user agent passed to every API call.
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
)

type payload map[string]interface{}

// Client provides access to the FIXR API methods.
type Client struct {
	Email      string
	Password   string
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	MagicURL   string      `json:"magic_login_url"`
	AuthToken  string      `json:"auth_token"`
	Error      string      `json:"message,omitempty"`
	StripeUser *stripeUser `json:"stripe_user"`
	httpClient *http.Client
}

// Event contains the event details for given event ID.
type Event struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Tickets []Ticket `json:"tickets"`
	Error   string   `json:"detail,omitempty"`
}

// Ticket contains all the information pertaining to a specific ticket for an event.
type Ticket struct {
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

// PromoCode contains the details of a specific promotional code.
type PromoCode struct {
	Code       string  `json:"code"`
	Price      float64 `json:"price"`
	BookingFee float64 `json:"booking_fee"`
	Currency   string  `json:"currency"`
	Max        int     `json:"max_per_user"`
	Remaining  int     `json:"remaining"`
	Error      string  `json:"message,omitempty"`
}

// Booking contains the resultant booking information.
type Booking struct {
	Event Event  `json:"event"`
	Name  string `json:"user_full_name"`
	PDF   string `json:"pdf"`
	State int    `json:"state"`
	Error string `json:"message,omitempty"`
}

// NewClient returns a FIXR client with the given email and password.
func NewClient(email, pass string) *Client {
	return &Client{Email: email, Password: pass, httpClient: new(http.Client)}
}

func (c *Client) get(addr string, auth bool, obj interface{}) error {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return errors.New("error creating GET request")
	}
	return c.req(req, auth, obj)
}

func (c *Client) post(addr string, val payload, auth bool, obj interface{}) error {
	data := new(bytes.Buffer)
	if addr == cardURL {
		pl := url.Values{}
		for key, value := range val {
			valueStr, ok := value.(string)
			if !ok {
				return errors.New("failed to build payload")
			}
			pl.Set(key, valueStr)
		}
		data.WriteString(pl.Encode())
	} else {
		if err := json.NewEncoder(data).Encode(val); err != nil {
			return errors.Wrap(err, "error jsonifying payload")
		}
	}
	req, err := http.NewRequest("POST", addr, data)
	if err != nil {
		return errors.New("error creating POST request")
	}
	return c.req(req, auth, obj)
}

func (c *Client) req(req *http.Request, auth bool, obj interface{}) error {
	req.Header.Set("User-Agent", UserAgent)
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthToken))
	}
	if req.URL.String() == cardURL {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", "application/json")
		// The following circumvents canonical formatting
		req.Header["FIXR-Platform"] = []string{"web"}
		req.Header["FIXR-Platform-Version"] = []string{FixrPlatformVer}
		req.Header["FIXR-App-Version"] = []string{FixrVersion}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error executing request")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// Logon authenticates the client with FIXR and returns an error if encountered.
func (c *Client) Logon() error {
	pl := payload{
		"email":    c.Email,
		"password": c.Password,
	}
	if err := c.post(loginURL, pl, false, c); err != nil {
		return errors.Wrap(err, "error logging on")
	}
	if len(c.Error) > 0 {
		return fmt.Errorf("error logging on: %s", c.Error)
	}
	return nil
}

// Event returns the event information for a given event ID (integer).
// An error will be returned if one is encountered.
func (c *Client) Event(id int) (*Event, error) {
	event := new(Event)
	if err := c.get(fmt.Sprintf(eventURL, id), false, event); err != nil {
		return nil, errors.Wrap(err, "error getting event")
	}
	if len(event.Error) > 0 {
		return nil, fmt.Errorf("error getting event: %s", event.Error)
	}
	return event, nil
}

// Promo checks for the existence of a promotion code for a given ticket ID.
// The returned *PromoCode can subsequently be passed to Book().
// An error will be returned if one is encountered.
func (c *Client) Promo(ticketID int, code string) (*PromoCode, error) {
	promo := new(PromoCode)
	if err := c.get(fmt.Sprintf(promoURL, ticketID, code), true, promo); err != nil {
		return nil, errors.Wrap(err, "error getting promo code")
	}
	if len(promo.Error) > 0 {
		return nil, fmt.Errorf("error getting promo code: %s", promo.Error)
	}
	return promo, nil
}

// Book books a ticket, given a *Ticket and an amout (with the option of a promo code).
// The booking details and an error, if encountered, will be returned.
func (c *Client) Book(ticket *Ticket, amount int, promo *PromoCode) (*Booking, error) {
	booking := new(Booking)
	pl := payload{
		"ticket_id": ticket.ID,
		"amount":    amount,
	}
	for t, msg := range map[bool]string{
		ticket.SoldOut: "ticket selection has sold out",
		ticket.Expired: "ticket selection has expired",
		/* ticket.Invalid: "ticket selection is invalid",
		Invalid can change upon ticket release (i.e. is time dependent),
		it should therefore be checked with an API call. */
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
	if err := c.post(bookingURL, pl, true, booking); err != nil {
		return nil, errors.Wrap(err, "error booking ticket")
	}
	if len(booking.Error) > 0 {
		return nil, fmt.Errorf("error booking ticket: %s", booking.Error)
	}
	return booking, nil
}
