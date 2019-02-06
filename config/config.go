package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"reflect"
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
	// Trust incoming certificates signed from unknown CAs
	TrustUnknownCAs bool `json:"trust_unknown_cas"`
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

	// print values
	rvc := reflect.ValueOf(*cfg)

	for i := 0; i < rvc.NumField(); i++ {

		fl := rvc.Type().Field(i)

		log.Infof("Config Field: %v has been successfully initialized with value: %v", fl.Name, rvc.Field(i).Interface())

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
