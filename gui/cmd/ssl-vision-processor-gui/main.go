package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/RoboCup-SSL/ssl-vision-processor/gui/internal/logging"
)

var address = flag.String("address", ":8085", "The address on which the vision processor GUI and API are served, default: :8085")
var visionAddress = flag.String("visionAddress", "224.5.23.2:10006", "The multicast address of field vision, default: 224.5.23.2:10006")
var skipInterfaces = flag.String("skipInterfaces", "", "Comma separated list of interface names to ignore when receiving multicast packets")
var logLevelFlag = flag.String("logLevel", "Info", "Log Level: Debug, Info, Warn, Error. Default: Info")
var logFile = flag.String("logFile", "logs/vision-processor-gui.log", "Rotating log file to write alongside stderr, empty to disable. Default: logs/vision-processor-gui.log")

func main() {
	flag.Parse()

	// Parsed before Setup so that the level is known when the handler is built,
	// but reported after it, so the warning goes through the same handler as
	// everything else.
	level, levelErr := logging.ParseLevel(*logLevelFlag)

	closeLog, logErr := logging.Setup(logging.Config{
		Level: level,
		File:  *logFile,
	})
	defer closeLog()

	if levelErr != nil {
		slog.Warn("falling back to info level", "err", levelErr)
	}

	if logErr != nil {
		slog.Warn("file logging disabled", "err", logErr)
	}

	slog.Info("Init VP GUI")

	// register Ctrl+C handler
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	slog.Info("UI is available", "url", formattedAddress())

	<-ctx.Done()

	// Restore default signal handling: while shutdown is in progress a second
	// Ctrl-C should kill the process outright rather than being swallowed.
	stopSignals()

	slog.Info("shutting down")

	// Cleanup of the HTTP server, receivers and bus clients goes here, bounded
	// by its own context so a wedged component cannot hang the process:
	//
	//	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//	defer cancel()
	//	httpServer.Shutdown(shutdownCtx)

	slog.Info("shutdown complete")
}

func formattedAddress() string {
	if strings.HasPrefix(*address, ":") {
		return "http://localhost" + *address
	}

	return "http://" + *address
}
