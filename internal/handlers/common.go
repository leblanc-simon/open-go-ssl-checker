package handlers

import (
	"leblanc.io/open-go-ssl-checker/internal/checker"
	"leblanc.io/open-go-ssl-checker/internal/store"
)

type AppContext struct {
	Store   *store.Store
	Checker *checker.CertificateService
}
