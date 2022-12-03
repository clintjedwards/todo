package service

import (
	"os"

	"github.com/clintjedwards/todo/internal/app"
	"github.com/clintjedwards/todo/internal/cli/cl"
	"github.com/clintjedwards/todo/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var cmdServiceStart = &cobra.Command{
	Use:   "start",
	Short: "Start the Todo GRPC/HTTP combined server",
	Long: `Start the Todo GRPC/HTTP combined server.

Todo runs as a GRPC backend combined with GRPC-WEB/HTTP. Running this command attempts to start the long
running service. This command will block and only gracefully stop on SIGINT or SIGTERM signals`,
	RunE: serverStart,
}

func init() {
	CmdService.AddCommand(cmdServiceStart)
}

func serverStart(cmd *cobra.Command, _ []string) error {
	cl.State.Fmt.Finish()

	configPath, _ := cmd.Flags().GetString("config")
	conf, err := config.InitAPIConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("error in config initialization")
	}

	setupLogging(conf.LogLevel, conf.DevMode)
	app.StartServices(conf)

	return nil
}

func setupLogging(loglevel string, pretty bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	zerolog.SetGlobalLevel(parseLogLevel(loglevel))
	if pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func parseLogLevel(loglevel string) zerolog.Level {
	switch loglevel {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		log.Error().Msgf("loglevel %s not recognized; defaulting to debug", loglevel)
		return zerolog.DebugLevel
	}
}
