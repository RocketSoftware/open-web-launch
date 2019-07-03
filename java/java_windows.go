package java

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

var registryKey = `Software\Rocket Software\Open Web Launch`

func getStringValueFromRootKey(rootKey registry.Key, key string) (string, error) {
	registryKey, err := registry.OpenKey(rootKey, registryKey, registry.QUERY_VALUE)
	defer registryKey.Close()
	if err != nil {
		return "", err
	}
	value, _, err := registryKey.GetStringValue(key)
	if err != nil {
		return "", err
	}
	return value, nil
}

func getUInt64ValueFromRootKey(rootKey registry.Key, key string) (uint64, error) {
	registryKey, err := registry.OpenKey(rootKey, registryKey, registry.QUERY_VALUE)
	defer registryKey.Close()
	if err != nil {
		return 0, err
	}
	value, _, err := registryKey.GetIntegerValue(key)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func getJavaExecutableFromRootKey(rootKey registry.Key) (string, error) {
	return getStringValueFromRootKey(rootKey, "Java")
}

func getJavaDirFromRootKey(rootKey registry.Key) (string, error) {
	return getStringValueFromRootKey(rootKey, "JavaDir")
}

func getJavaDetectionStrategyFromRootKey(rootKey registry.Key) (string, error) {
	return getStringValueFromRootKey(rootKey, "JavaDetection")
}

func getShowConsoleSettingFromRootKey(rootKey registry.Key) (uint64, error) {
	return getUInt64ValueFromRootKey(rootKey, "ShowConsole")
}

func getAddAppToControlPanelSettingFromRootKey(rootKey registry.Key) (uint64, error) {
	return getUInt64ValueFromRootKey(rootKey, "AddToControlPanel")
}

func getDisableVerificationSettingFromRootKey(rootKey registry.Key) (uint64, error) {
	return getUInt64ValueFromRootKey(rootKey, "DisableVerification")
}

func getJavaDetectionStrategy() string {
	strategy, err := getJavaDetectionStrategyFromRootKey(registry.CURRENT_USER)
	if err != nil {
		strategy, err = getJavaDetectionStrategyFromRootKey(registry.LOCAL_MACHINE)
	}
	if err != nil {
		strategy = ""
	}
	return strategy
}

func getDisableVerificationSetting() bool {
	disableVerification, err := getDisableVerificationSettingFromRootKey(registry.CURRENT_USER)
	if err != nil {
		disableVerification, err = getDisableVerificationSettingFromRootKey(registry.LOCAL_MACHINE)
	}
	if err != nil {
		return false
	}
	return disableVerification == 1
}

func getShowConsoleSetting() bool {
	showConsole, err := getShowConsoleSettingFromRootKey(registry.CURRENT_USER)
	if err != nil {
		showConsole, err = getShowConsoleSettingFromRootKey(registry.LOCAL_MACHINE)
	}
	if err != nil {
		return false
	}
	return showConsole == 1
}

func getJavaDirFromRegistry() (string, error) {
	javaSource = `Windows Registry - CURRENT_USER\` + registryKey + `\JavaDir`
	javadir, err := getJavaDirFromRootKey(registry.CURRENT_USER)
	if err != nil {
		javaSource = `Windows Registry - LOCAL_MACHINE\` + registryKey + `\JavaDir`
		return getJavaDirFromRootKey(registry.LOCAL_MACHINE)
	}
	return javadir, nil
}

func getJavaExecutableFromRegistry() (string, error) {
	javaSource = `Windows Registry - CURRENT_USER\` + registryKey + `\Java`
	java, err := getJavaExecutableFromRootKey(registry.CURRENT_USER)
	if err != nil {
		javaSource = `Windows Registry - LOCAL_MACHINE\` + registryKey + `\Java`
		return getJavaExecutableFromRootKey(registry.LOCAL_MACHINE)
	}
	return java, nil
}

func getDefaultJava(showConsole bool) string {
	javaSource = "PATH environment variable"
	if showConsole {
		return "java.exe"
	}
	return "javaw.exe"
}

func getJavaExecutable() string {
	showConsole := getShowConsoleSetting()
	strategy := getJavaDetectionStrategy()
	if strategy == "JavaHome" {
		java, err := getJavaExecutableUsingJavaHome(showConsole)
		if err != nil {
			return getDefaultJava(showConsole)
		}
		return java
	}
	javadir, err := getJavaDirFromRegistry()
	if err == nil {
		if showConsole {
			return filepath.Join(javadir, "bin", "java.exe")
		} else {
			return filepath.Join(javadir, "bin", "javaw.exe")
		}
	}
	java, err := getJavaExecutableFromRegistry()
	if err != nil {
		return getDefaultJava(showConsole)
	}
	return java
}

func getJavaExecutableUsingJavaDir(dir string) string {
	showConsole := getShowConsoleSetting()
	if showConsole {
		return filepath.Join(dir, "bin", "java.exe")
	}
	return filepath.Join(dir, "bin", "javaw.exe")
}

func getJARSignerExecutable() string {
	if filepath.IsAbs(javaExecutable) {
		return filepath.Join(filepath.Dir(javaExecutable), "jarsigner.exe")
	}
	return "jarsigner.exe"
}

func getJARSignerExecutableUsingJavaDir(dir string) string {
	return filepath.Join(dir, "bin", "jarsigner.exe")
}

func getAddAppToControlPanelSetting() bool {
	addAppToControlPanel, err := getAddAppToControlPanelSettingFromRootKey(registry.CURRENT_USER)
	if err != nil {
		addAppToControlPanel, err = getAddAppToControlPanelSettingFromRootKey(registry.LOCAL_MACHINE)
	}
	if err != nil {
		return false
	}
	return addAppToControlPanel == 1
}