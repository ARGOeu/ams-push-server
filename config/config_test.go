package config

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"strings"
	"testing"
)

type ConfigTestSuite struct {
	suite.Suite
}

// TestValidateRequired tests that the validateRequired method behaves as expected
func (suite *ConfigTestSuite) TestValidateRequired() {

	cfg := new(Config)
	e1 := cfg.validateRequired()

	// test the case where one the required field is not set
	suite.Equal("Empty value for field service_port", e1.Error())

	cfg2 := Config{
		ServicePort:               9000,
		Certificate:               "/path/cert.pem",
		CertificateKey:            "/path/certkey.pem",
		CertificateAuthoritiesDir: "/path/to/cas",
		AmsToken:                  "some_token",
		AmsHost:                   "example.come",
		AmsPort:                   8080,
		VerifySSL:                 false,
		TLSEnabled:                true,
		TrustUnknownCAs:           false,
		LogLevel:                  "INFO",
		SkipSubsLoad:              true,
		ACL:                       []string{"OU=my.local,O=mkcert development certificate"},
	}

	// test the case where where everything is set properly
	e2 := cfg2.validateRequired()
	suite.Nil(e2)
}

func (suite *ConfigTestSuite) TestLoadFromJson() {

	testCfg := `
{
  "service_port": 9000,
  "certificate": "/path/cert.pem",
  "certificate_key": "/path/certkey.pem",
  "certificate_authorities_dir": "/path/to/cas",
  "ams_token": "sometoken",
  "ams_host": "localhost",
  "ams_port": 8080,
  "verify_ssl": true,
  "tls_enabled": false,
  "trust_unknown_cas": true,
  "log_level": "INFO",
  "skip_subs_load": true,
  "acl": ["OU=my.local,O=mkcert development certificate"]
}
`
	cfg := new(Config)
	e1 := cfg.LoadFromJson(strings.NewReader(testCfg))

	// test the normal case
	suite.Equal(9000, cfg.ServicePort)
	suite.Equal("/path/cert.pem", cfg.Certificate)
	suite.Equal("/path/certkey.pem", cfg.CertificateKey)
	suite.Equal("/path/to/cas", cfg.CertificateAuthoritiesDir)
	suite.Equal("sometoken", cfg.AmsToken)
	suite.Equal("localhost", cfg.AmsHost)
	suite.Equal(8080, cfg.AmsPort)
	suite.Equal(true, cfg.VerifySSL)
	suite.Equal(false, cfg.TLSEnabled)
	suite.Equal(true, cfg.TrustUnknownCAs)
	suite.Equal("INFO", cfg.LogLevel)
	suite.Equal(true, cfg.SkipSubsLoad)
	suite.Equal([]string{"OU=my.local,O=mkcert development certificate"}, cfg.ACL)

	suite.Nil(e1)

	testCfg2 := `
{
  "service_port": 9000,
  "certificate": "/path/cert.pem",
  "certificate_key": "/path/certkey.pem",
  "certificate_authorities_dir": "/path/to/cas",
  "ams_token": "sometoken",
  "ams_host": "localhost",
  "ams_port": 8080,
  "verify_ssl": true,
  "trust_unknown_cas": true
` // miss closing bracket

	cfg2 := new(Config)
	e2 := cfg2.LoadFromJson(strings.NewReader(testCfg2))

	// test the case where the input is a malformed json
	suite.Equal("unexpected EOF", e2.Error())

	testCfg3 := `
{
  "service_port": 9000,
  "certificate": "/path/cert.pem",
  "certificate_key": "/path/certkey.pem",
  "certificate_authorities_dir": "/path/to/cas",
  "ams_token": "sometoken",
  "ams_host": "localhost",
  "ams_port": 8080,
  "verify_ssl": true,
  "tls_enabled": false,
  "trust_unknown_cas": true,
  "log_level": "unknown"
}
`

	cfg3 := new(Config)
	e3 := cfg3.LoadFromJson(strings.NewReader(testCfg3))
	// test the case where the log level is not one of the four wanted values
	suite.Equal("Invalid log level unknown", e3.Error())
}

func (suite *ConfigTestSuite) TestGetLogLevel() {

	cfg1 := new(Config)
	cfg1.LogLevel = "DEBUG"

	suite.Equal(log.DebugLevel, cfg1.GetLogLevel())

	cfg2 := new(Config)
	cfg2.LogLevel = "INFO"

	suite.Equal(log.InfoLevel, cfg2.GetLogLevel())

	cfg3 := new(Config)
	cfg3.LogLevel = "WARNING"

	suite.Equal(log.WarnLevel, cfg3.GetLogLevel())

	cfg4 := new(Config)
	cfg4.LogLevel = "ERROR"

	suite.Equal(log.ErrorLevel, cfg4.GetLogLevel())

	cfg5 := new(Config)
	cfg5.LogLevel = "unknown"

	suite.Equal(log.InfoLevel, cfg5.GetLogLevel())
}

func (suite *ConfigTestSuite) TestGetClientAuthType() {

	cfg1 := new(Config)
	cfg1.TrustUnknownCAs = true

	cfg2 := new(Config)
	cfg2.TrustUnknownCAs = false

	suite.Equal(tls.RequireAnyClientCert, cfg1.GetClientAuthType())
	suite.Equal(tls.RequireAndVerifyClientCert, cfg2.GetClientAuthType())
}

func TestConfigTestSuite(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	suite.Run(t, new(ConfigTestSuite))
}
