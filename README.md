# FIXR
A wrapper around FIXR's ticket API

## Installation
1. `go get github.com/pkg/errors`
2. `go get github.com/Nefarious-/fixr`
3. Done!

## Example Code:

```go
package main

import (
	"fmt"

	"github.com/Nefarious-/fixr"
)

func main() {
	fixr.Setup(true) // seed random number generator (true/false) & update FIXR version header
	c := fixr.NewClient("username", "password")
	if err := c.Logon(); err != nil {
		fmt.Printf("logon failed (%v)\n", err)
	}
	e, err := c.Event(141151926)
	if err != nil {
		fmt.Printf("failed getting event (%v)\n", err)
		return
	}
	for _, t := range e.Tickets {
		fmt.Printf("[%d] %s (£%.2f; Max: %d)\n", t.ID, t.Name, t.Price + t.BookingFee, t.Max)
	}
	b, err := c.Book(&e.Tickets[0], 1, nil)
	if err != nil {
		fmt.Printf("purchase failed (%v)", err)
		return
	}
	fmt.Printf("booked: %s (PDF: %s)\n", b.Event, b.PDF)
}
```
