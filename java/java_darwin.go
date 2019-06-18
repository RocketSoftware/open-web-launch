package java

import "path/filepath"

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
