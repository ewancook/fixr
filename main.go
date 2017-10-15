package main

import "fmt"

func main() {
	c := newClient("")
	c.logon("")
	fmt.Println(c.Email, c.FirstName, c.LastName, c.MagicURL, c.AuthToken)
	e := *new(event)
	fmt.Printf("%v \n %+v", c.getEvent(174206226, &e), e.Tickets[0])
}
