package handler

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting HTTPS handling for", r.Host)
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	targetHost, err := getHost(r, httpsPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Host = targetHost

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Printf("Hijack error: %v", err)
		return
	}
	defer func() {
		clientConn.Close()
		log.Println("Client connection closed")
	}()

	if _, err := clientConn.Write([]byte(successConnect)); err != nil {
		log.Printf("Failed to send 200 OK: %v", err)
		return
	}
	log.Println("Sent 200 Connection Established")

	targetHost = strings.Split(targetHost, ":")[0]
	log.Println("Generating certificate for:", targetHost)
	cert, err := h.generateCert(targetHost)
	if err != nil {
		log.Printf("Certificate generation failed: %v", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"http/1.1"},
	}

	log.Println("Performing TLS handshake with client")
	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("TLS handshake error: %v", err)
		return
	}
	defer tlsConn.Close()

	log.Println("Connecting to target server:", r.Host)
	targetConn, err := tls.Dial(
		"tcp", targetHost+":443", &tls.Config{
			ServerName:         targetHost,
			RootCAs:            h.caCertPool,
			InsecureSkipVerify: true,
		},
	)
	if err != nil {
		log.Printf("Target connection error: %v", err)
		return
	}
	defer targetConn.Close()

	// Туннелирование трафика
	log.Println("Starting tunneling")
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(targetConn, tlsConn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(tlsConn, targetConn)
		errChan <- err
	}()

	if err := <-errChan; err != nil {
		log.Printf("Tunnel error: %v", err)
	}
	log.Println("Tunneling complete")
}

func (h *Handler) generateCert(host string) (tls.Certificate, error) {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: host},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		template,
		h.caCert,
		&priv.PublicKey,
		h.caKey,
	)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		},
	)
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return tls.X509KeyPair(certPEM, keyPEM)
}
