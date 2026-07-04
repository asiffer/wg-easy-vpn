package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

const CONNECTION_ARG = "connection"
const SEPARATOR = string(os.PathSeparator)

var NO_COLOR bool = os.Getenv("NO_COLOR") != ""

var App = cli.Command{
	EnableShellCompletion: true,
	Commands:              []*cli.Command{&initCmd, &addCmd, &rmCmd},
	Suggest:               true,
}

func init() {
	// zerolog configuration
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05.000",
		NoColor:    NO_COLOR,
		FormatFieldName: func(i interface{}) string {
			if NO_COLOR {
				return fmt.Sprintf("%v=", i)
			}
			return color.RGB(255, 128, 0).Sprintf("%v=", i)
		},
		FormatLevel: func(i interface{}) string {
			levelStr := strings.ToUpper(i.(string))
			if NO_COLOR {
				return fmt.Sprintf("%5s", levelStr)
			}

			switch levelStr {
			case "DEBUG":
				// ANSI faint + gray
				return color.New(color.Faint).Sprintf("%5s", levelStr)
			case "INFO":
				return color.New(color.Bold).Sprintf("%5s", levelStr)
			case "WARN":
				return color.New(color.FgYellow).Sprintf("%5s", levelStr)
			case "ERROR":
				return color.New(color.BgRed).Add(color.FgWhite).Sprintf("%5s", levelStr)
			default:
				return fmt.Sprintf("%5s", levelStr)
			}
		},
	})
}

func ConfigurationInfo(raw string) (string, string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", "", fmt.Errorf("error while getting absolute path for %s: %w", raw, err)
	}

	// it means that we use the default wierguard path
	if abs == (SEPARATOR + raw) {
		return raw, path.Join(WIREGUARD_DIR, raw+CONFIG_SUFFIX), nil
	} else {
		name := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
		return name, abs, nil
	}
}
