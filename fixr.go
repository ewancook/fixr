package fixr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type client struct {
	Email      string
	Password   string
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MagicURL   string `json:"magic_login_url"`
	AuthToken  string `json:"auth_token"`
	httpClient *http.Client
}

type event struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Tickets []ticket `json:"tickets"`
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

func NewClient(email, pass string) *client {
	return &client{Email: email, Password: pass, httpClient: new(http.Client)}
}

func (c *client) req(method, addr string, val url.Values, auth bool, obj interface{}) error {
	data := bytes.NewBufferString(val.Encode())
	r, err := http.NewRequest(method, addr, data)
	if err != nil {
		return err
	}
	if auth {
		r.Header.Set("Authorization", fmt.Sprintf("Token %s", c.AuthToken))
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ContentLength = int64(data.Len())
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

func (c *client) Logon() error {
	pl := url.Values{}
	pl.Set("email", c.Email)
	pl.Set("password", c.Password)
	return c.req("POST", "https://api.fixr-app.com/api/v2/app/user/authenticate/with-email", pl, false, c)
}

func (c *client) Event(n int) (*event, error) {
	e := new(event)
	addr := fmt.Sprintf("https://api.fixr-app.com/api/v2/app/event/%d", n)
	return e, c.req("GET", addr, nil, false, e)
}
