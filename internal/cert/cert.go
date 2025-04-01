package cert

import (
	"crypto/tls"
)

func GenCert(tlsSertificate *tls.Certificate, hosts ...string) (*tls.Certificate, error) {}
