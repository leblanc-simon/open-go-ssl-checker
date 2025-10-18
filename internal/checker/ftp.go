package checker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/textproto"
)

func FtpGetTlsCertificate(
	host, port string,
	allowInsecure bool,
	serverNameOverride string,
) (*x509.Certificate, string, error) {
	serverAddr := net.JoinHostPort(host, port)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, "", fmt.Errorf("FTP: unable to connect to %s: %w", serverAddr, err)
	}
	defer conn.Close()

	tp := textproto.NewConn(conn)
	defer tp.Close()

	var ip string
	remoteAddr := conn.RemoteAddr()
	if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
		ip = tcpAddr.IP.String()
	} else {
		ip = remoteAddr.String() // Fallback si ce n'est pas une adresse TCP
	}

	_, _, err = tp.ReadResponse(220) // Welcome message
	if err != nil {
		return nil, ip, fmt.Errorf(
			"FTP: error reading initial response from %s: %w",
			serverAddr,
			err,
		)
	}

	err = tp.PrintfLine("AUTH TLS")
	if err != nil {
		return nil, ip, fmt.Errorf("FTP: error sending AUTH TLS to %s: %w", serverAddr, err)
	}

	_, _, err = tp.ReadResponse(234) // Server ready for TLS
	if err != nil {
		return nil, ip, fmt.Errorf("FTP: server %s did not accept AUTH TLS: %w", serverAddr, err)
	}

	effectiveServerName := serverNameOverride
	if effectiveServerName == "" {
		hostOnly, _, splitErr := net.SplitHostPort(serverAddr)
		if splitErr != nil {
			// Fallback if serverAddr has no port (unlikely here but good to handle)
			hostOnly = serverAddr
		}

		effectiveServerName = hostOnly
	}

	tlsConfig := &tls.Config{
		ServerName:         effectiveServerName,
		InsecureSkipVerify: allowInsecure, // WARNING: Security risk
	}

	tlsConn := tls.Client(conn, tlsConfig)
	if err = tlsConn.Handshake(); err != nil {
		return nil, ip, fmt.Errorf("FTP: TLS negotiation failed with %s: %w", serverAddr, err)
	}
	defer tlsConn.Close() // The original defer on conn will close the underlying connection

	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, ip, fmt.Errorf("FTP: no TLS certificate presented by %s", serverAddr)
	}

	// Optionally: Properly close the FTP session after obtaining the certificate
	// Create a new textproto.Conn for the TLS connection
	secureTp := textproto.NewConn(
		tlsConn,
	) // Do not explicitly close secureTp as tlsConn is already managed
	// by the defer above. Closing tp earlier would also be an option, after AUTH TLS.

	err = secureTp.PrintfLine("QUIT")
	if err != nil {
		// Not critical for cert retrieval, but good to know
		fmt.Printf(
			"FTP: Warning - error sending QUIT to %s: %v\n",
			serverAddr,
			err,
		)
	} else {
		_, _, err = secureTp.ReadResponse(221) // Bye
		if err != nil {
			fmt.Printf("FTP: Warning - error reading QUIT response from %s: %v\n", serverAddr, err)
		}
	}
	// The defer tlsConn.Close() will take care of closing the connection.

	return certs[0], ip, nil
}
