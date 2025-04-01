package controller

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/tp_security/internal/cert"
	"github.com/tp_security/internal/config"
	"github.com/tp_security/internal/global_errors"
)

type HttpSender interface {
	Do(req *http.Request) (*http.Response, error)
}

type Proxy struct {
	client          HttpSender
	certCA          *tls.Certificate
	tlsServerConfig *tls.Config
}

func NewProxy(cfg *config.Config) (*Proxy, error) {
	handler := &Proxy{
		client: &http.Client{
			Timeout:   cfg.WriteTimeout,
			Transport: http.DefaultTransport,
		},
		tlsServerConfig: cfg.TLSConfig,
	}

	err := handler.initCA(cfg.CertFile, cfg.KeyFile)

	return handler, err
}

func (p *Proxy) initCA(certFile, keyFile string) error {
	certRaw, err := os.ReadFile(certFile)
	if err != nil {
		log.Println(err)

		return fmt.Errorf("could not read certificate file: %w", err)
	}

	keyRaw, err := os.ReadFile(keyFile)
	if err != nil {
		log.Println(err)

		return fmt.Errorf("could not read private key file: %w", err)
	}

	certX509, err := tls.X509KeyPair(certRaw, keyRaw)
	if err != nil {
		log.Println(err)

		return fmt.Errorf("could not parse private key: %w", err)
	}

	certX509.Leaf, err = x509.ParseCertificate(certX509.Certificate[0])
	if err != nil {
		log.Println(err)

		return fmt.Errorf("could not parse certificate: %w", err)
	}

	p.certCA = &certX509

	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		p.tunneling(w, r)

		return
	}
	r.Header.Del("Proxy-Connection")
	r.RequestURI = ""
	resp, err := p.client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *Proxy) tunneling(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		log.Println(err)
		http.Error(w, global_errors.IncorrectHost.Error(), http.StatusBadRequest)
	}

	hostCert, err := cert.GenCert(p.certCA, []string{host}...)
	if err != nil {
		log.Println(err)
		http.Error(w, global_errors.InternalError.Error(), http.StatusInternalServerError)

		return
	}

	p.tlsServerConfig.Certificates = append(p.tlsServerConfig.Certificates, *hostCert)

	clientConn, err := hijackClientConnection(w)
	if err != nil {
		return
	}

	go func() {
		tlsConn := tls.Server(clientConn, p.tlsServerConfig)
		defer func() {
			err = tlsConn.Close()
			if err != nil {
				log.Println(err)
			}
		}()

		connReader := bufio.NewReader(tlsConn)

		for {
			err = p.doOneExchangeReqResp(connReader, tlsConn, r.Host, r.RemoteAddr)
			if err != nil {
				break
			}
		}
	}()
}

func (p *Proxy) doOneExchangeReqResp(
	connReader *bufio.Reader,
	tlsConn *tls.Conn,
	host string,
	remoteAddr string,
) error {
	_, err := http.ReadRequest(connReader)
	if err != nil {
		return err
	}

	return nil
}
