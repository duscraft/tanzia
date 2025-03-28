package main

import "time"

type Bill struct {
	Label       string
	Amount      int
	BillingDate time.Time
}
