package main

import (
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/renju24/backend/internal/apiserver"
	"github.com/renju24/backend/internal/pkg/database"
	"github.com/rs/zerolog"
)

// MachineIP is useful for logs.
func determineMachineIP() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		panic(err)
	}
	if len(addrs) == 0 {
		panic("could not lookup host IP.")
	}
	return addrs[0]
}

var runMode string

func main() {
	flag.StringVar(&runMode, "mode", "dev", "режим запуска prod/dev")
	flag.Parse()

	var (
		machineIP = determineMachineIP()
		port      = ":8008"
	)
	gin.SetMode(gin.ReleaseMode)
	time.Local = time.UTC

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(os.Stdout).
		With().
		Str("machine", machineIP+port).
		Timestamp().
		Caller().
		Logger().
		Level(zerolog.DebugLevel)

	log.SetOutput(logger)

	logger.Info().Msgf("Server is running in %q mode", runMode)

	logger.Info().Msg("Connecting to database.")
	db, err := database.New(os.Getenv("DATABASE_DSN"))
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer db.Close()

	// Init server.
	server := apiserver.NewAPIServer(runMode, db, gin.New(), &logger, db)

	logger.Info().Msg("Running api server")

	if err := server.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
