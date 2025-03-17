package main

import (
	"compress/flate"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"
)

func extract() {
	logger.Debugf("Attempting to extract \"%s\"", ipa)
	format := archiver.Zip{}
	directory = fileNameWithoutExtension(filepath.Base(ipa))

	if _, err := os.Stat(ipa); err != nil {
		logger.Errorf("Couldn't find \"%s\". Does it exist?", ipa)
		exit()
	}

	if _, err := os.Stat(directory); err == nil {
		logger.Debug("Detected previously extracted directory, cleaning it up...")

		err := os.RemoveAll(directory)
		if err != nil {
			logger.Errorf("Failed to clean up previously extracted directory: %s", err)
			exit()
		}

		logger.Info("Previously extracted directory cleaned up. ")
	}

	err := format.Unarchive(ipa, directory)
	if err != nil {
		logger.Errorf("Failed to extract %s: **%v**", ipa, err)
		os.Exit(1)
	}

	logger.Infof("Successfully extracted to \"%s\"", directory)
}

func archive() {
	logger.Debugf("Preparing to create ipa at \"%s\"", output)

	format := archiver.Zip{CompressionLevel: flate.BestCompression}
	zip := directory + ".zip"

	if _, err := os.Stat(zip); err == nil {
		logger.Debug("Detected previous archive, cleaning it up...")

		err := os.Remove(zip)
		if err != nil {
			logger.Errorf("Failed to clean up previous archive: %s", err)
			exit()
		}

		logger.Info("Previous archive cleaned up.")
	}

	logger.Debugf("Creating temporary archive from \"%s\"", directory)
	err := format.Archive([]string{filepath.Join(directory, "Payload")}, zip)
	if err != nil {
		logger.Errorf("Failed to create archive: %v", err)
		exit()
	}

	// Check if output file already exists and remove it if necessary
	if _, err := os.Stat(output); err == nil {
		logger.Debugf("Detected existing file at output path, cleaning it up...")

		err := os.Remove(output)
		if err != nil {
			logger.Errorf("Failed to clean up existing output file: %s", err)
			exit()
		}

		logger.Info("Existing output file cleaned up.")
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Errorf("Failed to create output directory: %v", err)
		exit()
	}

	err = os.Rename(zip, output)
	if err != nil {
		logger.Errorf("Failed to create output file \"%s\": %v", output, err)
		exit()
	}

	logger.Infof("Successfully created \"%s\"", output)
}
