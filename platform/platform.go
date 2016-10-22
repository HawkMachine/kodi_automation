package platform

import (
	"github.com/HawkMachine/kodi_automation/platform/cron"
)

type KodiConfig struct {
	Address  string
	Username string
	Password string
}

type TransmissionConfig struct {
	Address  string
	Username string
	Password string
}

type Config struct {
	Transmission TransmissionConfig
	Kodi         KodiConfig
}

func NewConfigFromStrings(
	kodiAddress, kodiUsername, kodiPassword,
	trAddress, trUsername, trPassword string) Config {
	return Config{
		Transmission: TransmissionConfig{
			Address:  trAddress,
			Username: trUsername,
			Password: trPassword,
		},
		Kodi: KodiConfig{
			Address:  kodiAddress,
			Username: kodiUsername,
			Password: kodiPassword,
		},
	}
}

type Platform struct {
	Cron   *cron.Cron
	Config Config
}

func New(c Config) *Platform {
	return &Platform{
		Cron:   &cron.Cron{},
		Config: c,
	}
}
