package uploader

import (
	"github.com/natemarks/secret-hoard/tools"

)

// CSVProcessor is an interface defining a method for handling data.
type CSVProcessor interface {
	Process(cfg tools.Config, log *tools.Logger)
}
