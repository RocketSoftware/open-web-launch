package settings

import (
	"bytes"
	"howett.net/plist"
	"io/ioutil"
	"path/filepath"
)

const (
	systemSettingsFile = "/Library/Preferences/com.rs.openweblaunch.plist"
)

type Settings struct {
	DisableVerification           bool `plist:"DisableVerification"`
	DisableVerificationSameOrigin bool `plist:"DisableVerificationSameOrigin"`
}

func getJavaExecutable() string {
	if java, err := getJavaExecutableUsingJavaHome(true); err == nil {
		return java
	}
	javaSource = "PATH environment variable"
	return "java"
}

func getJavaExecutableUsingJavaDir(dir string) string {
	return filepath.Join(dir, "bin", "java")
}

func getJARSignerExecutable() string {
	return "jarsigner"
}

func getJARSignerExecutableUsingJavaDir(dir string) string {
	return filepath.Join(dir, "bin", "jarsigner")
}

func getDisableVerificationSetting() bool {
	settings, err := decodeSettings()
	if err != nil {
		return false
	}
	return settings.DisableVerification
}

func getDisableVerificationSameOriginSetting() bool {
	settings, err := decodeSettings()
	if err != nil {
		return false
	}
	return settings.DisableVerificationSameOrigin
}

func getAddAppToControlPanelSetting() bool {
	return false
}

func decodeSettings() (*Settings, error) {
	var settings Settings
	data, err := ioutil.ReadFile(systemSettingsFile)
	if err != nil {
		return nil, err
	}
	decoder := plist.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}
