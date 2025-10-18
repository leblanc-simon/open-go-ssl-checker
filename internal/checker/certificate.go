package checker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"leblanc.io/open-go-ssl-checker/internal/store"
	"leblanc.io/open-go-ssl-checker/internal/types"
	"leblanc.io/open-go-ssl-checker/internal/websocket"
)

type CertificateService struct {
	Store *store.Store
	Hub   *websocket.Hub
}

func NewCertificateService(s *store.Store, h *websocket.Hub) *CertificateService {
	return &CertificateService{Store: s, Hub: h}
}

func (cs *CertificateService) CheckAndStoreCertificate(
	projectID, host, port, projectType string,
	allowInsecure bool,
) {
	log.Printf(
		"Checking certificate for project %s (%s:%s, type: %s)\n",
		projectID,
		host,
		port,
		projectType,
	)

	var cert *x509.Certificate
	var ip string

	var err error

	switch projectType {
	case "ftp":
		cert, ip, err = FtpGetTlsCertificate(host, port, allowInsecure, host)
		if err != nil {
			log.Printf(
				"Error retrieving FTP certificate for %s:%s: %v",
				host,
				port,
				err,
			)
			cs.recordCheckFailure(projectID, fmt.Sprintf("FTP retrieval: %v", err))

			return
		}

	default:
		conn, dialErr := tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{
			InsecureSkipVerify: allowInsecure,
			ServerName:         host,
		})
		if dialErr != nil {
			log.Printf("Error connecting via TLS to %s:%s: %v", host, port, dialErr)
			cs.recordCheckFailure(projectID, fmt.Sprintf("TLS connection: %v", dialErr))

			return
		}
		defer conn.Close()

		remoteAddr := conn.RemoteAddr()
		if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
			ip = tcpAddr.IP.String()
		} else {
			ip = remoteAddr.String()
		}

		if len(conn.ConnectionState().PeerCertificates) == 0 {
			log.Printf("No certificate presented by %s:%s", host, port)
			cs.recordCheckFailure(projectID, "No certificate presented")

			return
		}

		cert = conn.ConnectionState().PeerCertificates[0]
	}

	if cert == nil {
		log.Printf("Certificate not retrieved for project %s (%s:%s)", projectID, host, port)
		cs.recordCheckFailure(projectID, "Certificate not retrieved (generic)")

		return
	}

	cs.handleCertificateInfo(projectID, cert, ip)
}

func (cs *CertificateService) handleCertificateInfo(projectID string, cert *x509.Certificate, ip string) {
	var domains []string
	if len(cert.DNSNames) > 0 {
		domains = cert.DNSNames
	} else {
		domains = append(domains, cert.Subject.CommonName)
	}

	expiryDate := cert.NotAfter
	daysRemaining := int(time.Until(expiryDate).Hours() / 24)
	issuer := cert.Issuer.String()

	projectName, err := cs.Store.GetProjectName(projectID)
	if err != nil {
		log.Printf(
			"Warning: Unable to retrieve the project name %s for storing the check: %v",
			projectID,
			err,
		)

		projectName = "Unknown"
	}

	checkData := types.CertificateCheck{
		CheckTime:     time.Now(),
		ProjectID:     projectID,
		ProjectName:   projectName,
		Domains:       strings.Join(domains, ", "),
		IP:            ip,
		Issuer:        issuer,
		ExpiryDate:    expiryDate.Format("2006-01-02"),
		DaysRemaining: daysRemaining,
	}

	if err := cs.Store.AddCertificateCheck(checkData); err != nil {
		log.Printf(
			"Error inserting verification data for project %s: %v",
			projectID,
			err,
		)
	} else {
		log.Printf("Certificate verification stored for project %s. Domains: %s, Expires on: %s (%d days remaining)",
			projectID, checkData.Domains, checkData.ExpiryDate, checkData.DaysRemaining)

		if cs.Hub != nil {
			cs.Hub.NotifyUpdate()
		}
	}
}

func (cs *CertificateService) recordCheckFailure(projectID string, failureReason string) {
	projectName, err := cs.Store.GetProjectName(projectID)
	if err != nil {
		projectName = "Unknown"
	}

	checkData := types.CertificateCheck{
		CheckTime:     time.Now(),
		ProjectID:     projectID,
		ProjectName:   projectName,
		Domains:       fmt.Sprintf("Failure: %s", failureReason),
		ExpiryDate:    "N/A",
		DaysRemaining: -1,
	}
	if err := cs.Store.AddCertificateCheck(checkData); err != nil {
		log.Printf(
			"Error recording verification failure for project %s: %v",
			projectID,
			err,
		)
	} else {
		log.Printf("Verification failure recorded for project %s: %s", projectID, failureReason)

		if cs.Hub != nil {
			cs.Hub.NotifyUpdate()
		}
	}
}
