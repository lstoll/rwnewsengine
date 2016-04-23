package rwnewsengine

type Config struct {
	ReadabilityKey    string `env:"READABILITY_KEY,required"`
	GroupAddress      string `env:"GROUP_EMAIL_ADDRESS,required"`
	MailgunDomain     string `env:"RWNEWS_MAILGUN_DOMAIN,required"`
	MailgunPrivateKey string `env:"MAILGUN_API_KEY,required"`
	MailgunPublicKey  string `env:"MAILGUN_PUBLIC_KEY,required"`
	Port              string `env:"PORT,default=5000"`
}
