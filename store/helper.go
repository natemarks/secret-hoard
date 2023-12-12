package store

import (
	"os"

	"github.com/rs/zerolog/log"
)

func readFileToString(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeStringToFile(filename, contents string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msgf("error closing file %s", filename)
		}
	}(file)

	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}
	return nil

}
