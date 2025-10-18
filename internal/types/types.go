package types

import (
	"time"
)

type Project struct {
	ID            string
	Name          string
	Host          string
	Port          string
	Type          string
	AllowInsecure bool
}

type CertificateCheck struct {
	ID            int64
	CheckTime     time.Time
	ProjectID     string
	ProjectName   string
	Domains       string
	IP            string
	Issuer 	      string
	ExpiryDate    string
	DaysRemaining int
}

type ProjectCheckSummary struct {
	ProjectID     string
	ProjectName   string
	Host          string
	Port          string
	Type          string
	AllowInsecure bool
	CheckTime     *time.Time
	Domains       string
	IP            string
	Issuer 	      string
	ExpiryDate    string
	DaysRemaining *int
}
