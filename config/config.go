package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Config contains all the needed information for the server to function properly
type Config struct {
	// Which port to bind to
	ServicePort int `json:"service_port" required:"true"`
	// Certificate file in order to enable TLS
	Certificate string `json:"certificate" required:"true"`
	// Certificate's private key
	CertificateKey string `json:"certificate_key" required:"true"`
	// Directory containing all the appropriate files to load the CA
	CertificateAuthoritiesDir string `json:"certificate_authorities_dir" required:"true"`
	// token for ams interaction
	AmsToken string `json:"ams_token" required:"true"`
	// Ams endpoint
	AmsHost string `json:"ams_host" required:"true"`
	// Ams port
	AmsPort int `json:"ams_port" required:"true"`
	// Whether or not any http client spawn inside  the service should accept to talk to non https connections
	VerifySSL bool `json:"verify_ssl"`
	// whether or not the service will start with tls enabled
	TLSEnabled bool `json:"tls_enabled"`
	// Trust incoming certificates signed from unknown CAs
	TrustUnknownCAs bool `json:"trust_unknown_cas"`
	// log level(DEBUG,INFO,WARNING,ERROR)
	LogLevel string `json:"log_level" required:"true"`
	// whether or not it should try to load any push enabled subscriptions, upon starting up
	SkipSubsLoad bool `json:"skip_subs_load"`
	// tls configuration to be used by the grpc server
	tlsConfig *tls.Config
}

var logLevels = map[string]log.Level{
	"DEBUG":   log.DebugLevel,
	"INFO":    log.InfoLevel,
	"WARNING": log.WarnLevel,
	"ERROR":   log.ErrorLevel,
}

// GetTLSConfig returns the tls configuration needed for the grpc server
func (cfg *Config) GetTLSConfig() *tls.Config {
	return cfg.tlsConfig
}

// GetLogLevel maps the provided string log level inside the config to a log.Level
// if it can't map it, it will return log.LevelInfo
func (cfg *Config) GetLogLevel() log.Level {

	logLevel, ok := logLevels[strings.ToUpper(cfg.LogLevel)]

	if !ok {
		return log.InfoLevel
	}

	return logLevel
}

// LoadFromJson fills the config struct with the contents of the reader
func (cfg *Config) LoadFromJson(from io.Reader) error {

	// load the configuration into the struct
	err := json.NewDecoder(from).Decode(cfg)
	if err != nil {
		return err
	}

	// check if all required fields are set
	err = cfg.validateRequired()
	if err != nil {
		return err
	}

	// check if the given log value is correct
	_, ok := logLevels[strings.ToUpper(cfg.LogLevel)]
	if !ok {
		return errors.Errorf("Invalid log level %v", cfg.LogLevel)
	}

	// print values
	rvc := reflect.ValueOf(*cfg)

	for i := 0; i < rvc.NumField(); i++ {

		fl := rvc.Type().Field(i)

		// Check if the field's exported otherwise .Interface() will panic
		if fl.PkgPath != "" {
			continue
		}

		// skip non configuration fields
		if fl.Tag.Get("json") == "" {
			continue
		}

		log.WithFields(
			log.Fields{
				"type":  "service_log",
				"field": fl.Tag.Get("json"),
				"value": rvc.Field(i).Interface(),
			},
		).Info("Configuration field has been successfully initialized")

	}

	// load tls configuration
	if cfg.TLSEnabled {
		err = cfg.loadTLSConfig()
		if err != nil {
			return err
		}
	}

	return nil
}

// validateRequired accepts checks whether or not all required fields are set
func (cfg *Config) validateRequired() error {

	v := reflect.ValueOf(*cfg)

	for i := 0; i < v.NumField(); i++ {

		sf := v.Type().Field(i)

		// Check if the field's exported otherwise .Interface() will panic
		if sf.PkgPath != "" {
			continue
		}

		// check if the field has the required tag
		if sf.Tag.Get("required") != "true" {
			continue
		}

		fieldValue := v.Field(i).Interface()
		zeroFieldValue := reflect.Zero(reflect.TypeOf(v.Field(i).Interface())).Interface()

		// check if the field's value is equal to its zero value, it means that is not set
		if reflect.DeepEqual(fieldValue, zeroFieldValue) {
			return errors.Errorf("Empty value for field %v", sf.Tag.Get("json"))
		}
	}

	return nil
}

// loadTLSConfig initialises the tls configuration field
func (cfg *Config) loadTLSConfig() error {

	c, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.CertificateKey)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		ClientAuth:               cfg.GetClientAuthType(),
		Certificates:             []tls.Certificate{c},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		ClientCAs:                cfg.loadCAs(),
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		NextProtos: []string{"h2"},
	}

	cfg.tlsConfig = tlsConfig
	return nil
}

// GetClientAuthType returns which client auth strategy should the server follow when validating a certificate
func (cfg *Config) GetClientAuthType() tls.ClientAuthType {

	authType := tls.RequireAndVerifyClientCert

	if cfg.TrustUnknownCAs {
		authType = tls.RequireAnyClientCert
	}

	return authType
}

// loadCAs walks the specified CertificateAuthoritiesDir and uses each .pem file to build the trusted CA pool
func (cfg *Config) loadCAs() *x509.CertPool {

	log.WithFields(
		log.Fields{
			"type": "service_log",
			"path": cfg.CertificateAuthoritiesDir,
		},
	).Info("Trying to load CAs")

	pattern := "*.pem"
	roots := x509.NewCertPool()
	err := filepath.Walk(cfg.CertificateAuthoritiesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":  "error_log",
					"path":  path,
					"error": err.Error(),
				},
			).Error("Prevent panic by handling failure accessing a path")
			return err
		}
		if ok, _ := filepath.Match(pattern, info.Name()); ok {
			bytes, err := ioutil.ReadFile(filepath.Join(cfg.CertificateAuthoritiesDir, info.Name()))
			if err != nil {
				return err
			}
			if ok = roots.AppendCertsFromPEM(bytes); !ok {
				return errors.New("Something went wrong while parsing certificate: " + filepath.Join(cfg.CertificateAuthoritiesDir, info.Name()))
			}
		}
		return nil
	})

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":  "error_log",
				"path":  cfg.CertificateAuthoritiesDir,
				"error": err.Error(),
			},
		).Error("Could not walk down the path")
	} else {
		log.WithFields(
			log.Fields{
				"type": "service_log",
			},
		).Info("All CAs parsed and loaded successfully")
	}

	return roots
}
