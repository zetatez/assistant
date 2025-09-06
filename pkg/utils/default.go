package utils

import "fmt"

func GetOSDefault(objType string) (string, error) {
	ObjTypeMap := map[string]func() (string, error){
		"shell":    GetOSDefaultShell,
		"terminal": GetOSDefaultTerminal,
		"editor":   GetOSDefaultEditor,
		"browser":  GetOSDefaultBrowser,
	}
	if fn, ok := ObjTypeMap[objType]; ok {
		return fn()
	}
	return "", fmt.Errorf("Unsupported ObjType")
}

func GetOSDefaultShell() (string, error) {
	OsDefault := map[string]string{
		OsMap["linux"]: "sh",
		OsMap["macos"]: "sh",
	}
	if val, ok := OsDefault[GetOSType()]; ok {
		return val, nil
	}
	return "", fmt.Errorf("Unsupported OS")
}

func GetOSDefaultTerminal() (string, error) {
	OsDefault := map[string]string{
		OsMap["linux"]: "st",
		OsMap["macos"]: "kitty",
	}
	if val, ok := OsDefault[GetOSType()]; ok {
		return val, nil
	}
	return "", fmt.Errorf("Unsupported OS")
}

func GetOSDefaultEditor() (string, error) {
	OsDefault := map[string]string{
		OsMap["linux"]: "nvim",
		OsMap["macos"]: "nvim",
	}
	if val, ok := OsDefault[GetOSType()]; ok {
		return val, nil
	}
	return "", fmt.Errorf("Unsupported OS")
}

func GetOSDefaultBrowser() (string, error) {
	OsDefault := map[string]string{
		OsMap["linux"]: "chrome",
		OsMap["macos"]: "chrome",
	}
	if val, ok := OsDefault[GetOSType()]; ok {
		return val, nil
	}
	return "", fmt.Errorf("Unsupported OS")
}
