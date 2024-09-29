package session

import (
	"embed"
	"encoding/json"
	"math/rand"

	"github.com/gotd/td/telegram"
)

type Devices struct {
	Tdesktop struct {
		ApiID          int      `json:"api_id"`
		ApiHash        string   `json:"api_hash"`
		SystemVersions []string `json:"system_versions"`
		LangPack       string   `json:"lang_pack"`
		DeviceModels   []string `json:"device_models"`
		AppVersion     string   `json:"app_version"`
		LangCode       string   `json:"lang_code"`
		SystemLangCode string   `json:"system_lang_code"`
	} `json:"tdesktop"`
}

type Device struct {
	SystemVersion  string `json:"system_version"`
	LangPack       string `json:"lang_pack"`
	DeviceModel    string `json:"device_model"`
	AppVersion     string `json:"app_version"`
	LangCode       string `json:"lang_code"`
	SystemLangCode string `json:"system_lang_code"`
}

//go:embed devices.json
var f embed.FS

func Windows() telegram.DeviceConfig {
	sessionJson, _ := f.ReadFile("devices.json")
	var sessionStruct Devices
	// not handling error because this file is embedded, so it will work.
	json.Unmarshal(sessionJson, &sessionStruct)

	systemVersion := sessionStruct.Tdesktop.SystemVersions[rand.Intn(len(sessionStruct.Tdesktop.SystemVersions))]
	config := telegram.DeviceConfig{
		DeviceModel:    sessionStruct.Tdesktop.DeviceModels[rand.Intn(len(sessionStruct.Tdesktop.DeviceModels))],
		SystemVersion:  systemVersion,
		AppVersion:     sessionStruct.Tdesktop.AppVersion,
		SystemLangCode: sessionStruct.Tdesktop.SystemLangCode,
		LangCode:       sessionStruct.Tdesktop.LangCode,
		LangPack:       sessionStruct.Tdesktop.LangPack,
	}
	return config
}
