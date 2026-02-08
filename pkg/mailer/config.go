package mailer

type Config struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	From      string `yaml:"from"`
	FromName  string `yaml:"fromName"`
	TLSPolicy string `yaml:"tlsPolicy"`
	AuthType  string `yaml:"authType"`
	Timeout   int    `yaml:"timeout"`
}

func DefaultConfig() Config {
	return Config{
		Port:      465,
		TLSPolicy: "TLS",
		AuthType:  "Plain",
		Timeout:   10,
	}
}
