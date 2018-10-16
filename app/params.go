package app

import (
	"flag"
	"os"
)

type ProcParams struct {
	ConfigPath  string
	Environment string
}

func LoadParams() (*ProcParams, error) {

	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	configPath := flagSet.String("config", "./conf", "config path")
	env := flagSet.String("env", "dev", "project Env")
	flagSet.Parse(os.Args[1:])

	params := &ProcParams{
		ConfigPath:  *configPath,
		Environment: *env,
	}

	return params, nil
}
