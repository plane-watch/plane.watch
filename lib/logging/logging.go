package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"runtime/pprof"
	"time"
)

var (
	isCli bool
)

const (
	VeryVerbose = "very-verbose"
	Debug       = "debug"
	Quiet       = "quiet"
	CpuProfile  = "cpu-profile"
)

func IncludeVerbosityFlags(app *cli.App) {
	app.Flags = append(app.Flags,
		&cli.BoolFlag{
			Name:  VeryVerbose,
			Usage: "Enable trace level debugging",
		},
		&cli.BoolFlag{
			Name:    Debug,
			Usage:   "Show Extra Debug Information",
			EnvVars: []string{"DEBUG"},
		},
		&cli.BoolFlag{
			Name:    Quiet,
			Usage:   "Only show important messages",
			EnvVars: []string{"QUIET"},
		},
		&cli.StringFlag{
			Name:  CpuProfile,
			Usage: "Specifying this parameter causes a CPU Profile to be generated",
		},
	)
	// append our after func to stop profiling
	if nil == app.After {
		app.After = StopProfiling
	} else {
		f := app.After
		app.After = func(c *cli.Context) error {
			err := f(c)
			_ = StopProfiling(c)
			return err
		}
	}
}

func SetLoggingLevel(c *cli.Context) {
	SetVerboseOrQuiet(
		c.Bool(VeryVerbose),
		c.Bool(Debug),
		c.Bool(Quiet),
	)
	if "" != c.String(CpuProfile) {
		ConfigureForProfiling(c.String(CpuProfile))
	}
}

func SetVerboseOrQuiet(trace, verbose, quiet bool) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if trace {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	//log.Info().Str("log-level", zerolog.GlobalLevel().String()).Msg("Logging Set")
}

func cliWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.UnixDate}
}

func ConfigureForCli() {
	isCli = true
	log.Logger = log.Output(cliWriter())
}

func ConfigureForProfiling(outFile string) {
	f, err := os.Create(outFile)
	if nil != err {
		panic(err)
	}
	err = pprof.StartCPUProfile(f)
	if nil != err {
		panic(err)
	}
}

func StopProfiling(c *cli.Context) error {
	if fileName := c.String(CpuProfile); "" != fileName {
		pprof.StopCPUProfile()
		println("To analyze the profile, use this cmd")
		println("go tool pprof -http=:7777", fileName)
	}
	return nil
}
