package job

import (
	"github.com/cebrains/jupiter/pkg/flag"
)

func init() {
	flag.Register(
		&flag.StringFlag{
			Name:    "job",
			Usage:   "--job",
			Default: "",
		},
	)
}

// Runner ...
type Runner interface {
	Run()
}
