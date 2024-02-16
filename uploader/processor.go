package uploader

import (
	"github.com/natemarks/secret-hoard/tools"
	"github.com/rs/zerolog"
)

// CSVProcessor is an interface defining a method for handling data.
type CSVProcessor interface {
	Process(cfg tools.Config, log *zerolog.Logger)
}
