package main

import "github.com/natemarks/secret-hoard/get"

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}
	log := cfg.GetLogger()
	log.Info().Msgf("config: %+v", cfg)
	get.DownloadSecret(cfg.SecretID, cfg.FilePath, &log)
}
