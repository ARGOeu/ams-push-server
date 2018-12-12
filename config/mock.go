package config

// NewMockConfig returns a config to be used in tests
func NewMockConfig() *Config {
	cfg := new(Config)
	cfg.ServicePort = 9000
	cfg.Certificate = "/path/cert.pem"
	cfg.CertificateKey = "/path/certkey.pem"
	cfg.CertificateAuthoritiesDir = "/path/to/cas"
	cfg.AmsToken = "sometoken"
	cfg.AmsHost = "localhost"
	cfg.AmsPort = 8080
	cfg.VerifySSL = true
	cfg.TrustUnknownCAs = false
	return cfg
}
