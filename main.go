package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    false,
	ReportTimestamp: true,
	TimeFormat:      time.TimeOnly,
	Level:           log.DebugLevel,
	Prefix:          "Patcher",
})

var (
	info      map[string]any
	directory string
	assets    string
	ipa       string
	output    string
)

func main() {
	app := &cli.App{
		Name:  "patcher-ios",
		Usage: "Patches the Discord ipa with icons, utilities and features to ease usability.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input path for the Discord ipa file to patch",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output path for the patched ipa file",
				Required: true,
			},
		},
		Action: func(context *cli.Context) error {
			ipa = context.String("input")
			output = context.String("output")

			if ipa == "" {
				logger.Error("Please provide a path to the input ipa using --input or -i flag.")
				os.Exit(1)
			}

			// Convert relative paths to absolute to ensure consistency
			if !filepath.IsAbs(ipa) {
				absPath, err := filepath.Abs(ipa)
				if err == nil {
					ipa = absPath
				}
			}

			if !filepath.IsAbs(output) {
				absPath, err := filepath.Abs(output)
				if err == nil {
					output = absPath
				}
			}

			logger.Infof("Requested ipa patch for \"%s\"", ipa)
			logger.Infof("Output will be saved to \"%s\"", output)

			extract()
			loadInfo()

			setReactNavigationName()
			setSupportedDevices()
			setFileAccess()
			setAppName()
			setIcons()
			setURLScheme()

			saveInfo()
			archive()

			exit()
			return nil
		},
	}

	assets = os.TempDir()

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
