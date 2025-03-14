package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
)

type Manifest struct {
	Metadata struct {
		Build         string `json:"build"`
		Commit        string `json:"commit"`
		ConfirmUpdate bool   `json:"confirm_update"`
	} `json:"metadata"`
	Hashes map[string]string `json:"hashes"`
}

func setIcons() {
	logger.Debug("Downloading app icons...")

	icons := filepath.Join(assets, "icons.zip")
	download("https://assets.unbound.rip/icons/ios.zip", icons)

	logger.Info("Downloaded app icons.")

	logger.Debug("Adding Unbound as alternate icon...")

	// Add UnboundIcon to iPhone alternate icons
	iPhoneIcons := info["CFBundleIcons"].(map[string]interface{})
	if iPhoneIcons["CFBundleAlternateIcons"] == nil {
		iPhoneIcons["CFBundleAlternateIcons"] = make(map[string]interface{})
	}
	alternateIcons := iPhoneIcons["CFBundleAlternateIcons"].(map[string]interface{})
	alternateIcons["UnboundIcon"] = map[string]interface{}{
		"CFBundleIconFiles": []string{"UnboundIcon60x60"},
	}

	// Add UnboundIcon to iPad alternate icons
	iPadIcons := info["CFBundleIcons~ipad"].(map[string]interface{})
	if iPadIcons["CFBundleAlternateIcons"] == nil {
		iPadIcons["CFBundleAlternateIcons"] = make(map[string]interface{})
	}
	alternateIconsIpad := iPadIcons["CFBundleAlternateIcons"].(map[string]interface{})
	alternateIconsIpad["UnboundIcon"] = map[string]interface{}{
		"CFBundleIconFiles": []string{"UnboundIcon60x60", "UnboundIcon76x76"},
	}

	zip := archiver.Zip{OverwriteExisting: true}
	discord := filepath.Join(directory, "Payload", "Discord.app")

	if err := zip.Unarchive(icons, discord); err == nil {
		logger.Info("Added Unbound as alternate app icon.")
	} else {
		logger.Errorf("Failed to add Unbound app icon: %v", err)
		exit()
	}
}

func setReactNavigationName() {
	logger.Debug("Attempting to rename React Navigation...")
	modulesPath := filepath.Join(directory, "Payload", "Discord.app", "assets", "_node_modules", ".pnpm")

	if exists, _ := exists(modulesPath); !exists {
		logger.Debug("React Navigation does not exist, no need to rename it.")
		return
	}

	manifestPath := filepath.Join(directory, "Payload", "Discord.app", "manifest.json")

	if exists, _ := exists(manifestPath); !exists {
		logger.Debug("React Navigation does not exist, no need to rename it.")
		return
	}

	content, err := os.ReadFile(manifestPath)

	if err != nil {
		logger.Errorf("Couldn't find manifest.json inside the ipa payload. %v", err)
		exit()
	}

	manifest := Manifest{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		logger.Errorf("Failed to unmarshal manifest.json. %v", err)
		exit()
	}

	if manifest.Hashes == nil {
		logger.Infof("No hashes found in manifest.json. Skipping React Navigation rename.")
		return
	}

	// Change manifest hash path
	for key := range manifest.Hashes {
		if !strings.Contains(key, "@react-navigation+elements") {
			continue
		}

		value := manifest.Hashes[key]
		split := strings.Split(key, "/")

		for idx, segment := range split {
			if !strings.Contains(segment, "@react-navigation+elements") {
				continue
			}

			split[idx] = "@react-navigation+elements@patched"
		}

		delete(manifest.Hashes, key)

		newKey := strings.Join(split, "/")
		manifest.Hashes[newKey] = value
	}

	content, err = json.Marshal(manifest)

	if err != nil {
		logger.Errorf("Failed to marshal modified manifest structure. %v", err)
		return
	}

	err = os.WriteFile(manifestPath, content, 0644)

	if err != nil {
		logger.Errorf("Failed to write modified manifest.json file. %v", err)
		return
	} else {
		logger.Info("Wrote modified manifest.json file.")
	}

	// Rename node_modules module folder(s)
	files, err := os.ReadDir(modulesPath)

	if err != nil {
		logger.Errorf("Failed to read node_modules directory. Skipping React Navigation rename. %v", err)
		return
	}

	directories := filter(files, func(entry fs.DirEntry) bool {
		return strings.Contains(entry.Name(), "@react-navigation+elements")
	})

	for _, directory := range directories {
		currentName := filepath.Join(modulesPath, directory.Name())
		newName := filepath.Join(modulesPath, "@react-navigation+elements@patched")

		if err := os.Rename(currentName, newName); err != nil {
			logger.Errorf("Failed to rename React Navigation directory: %v %v", directory.Name(), err)
		} else {
			logger.Infof("Renamed React Navigation directory: %v", directory.Name())
		}
	}

	logger.Info("Successfully renamed React Navigation directories.")
}
