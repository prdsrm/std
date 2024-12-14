package postgres

import (
	"database/sql"
)

type ChannelsPts struct {
	ChannelID int64
	Pts       int
}

type State struct {
	BotUserID int64         `db:"bot_user_id"`
	ID        sql.NullInt64 `db:"id"`
	Pts       sql.NullInt32 `db:"pts"`
	Qts       sql.NullInt32 `db:"qts"`
	Date      sql.NullInt32 `db:"date"`
	Seq       sql.NullInt32 `db:"seq"`
}

type BotAndEntityRelationship struct {
	FirstEntityID  int64         `db:"entity_id_1"`
	SecondEntityID int64         `db:"entity_id_2"`
	Hash           sql.NullInt64 `db:"hash"`
}

type Bot struct {
  UserID      int64  `db:"user_id" json:"user_id"`
  PhoneNumber string `db:"phone_number" json:"phone_number"`
  Username    string `db:"username" json:"username"`
  Password    string `db:"password" json:"password"`
  Title       string `db:"title" json:"title"`
  Premium     bool   `db:"premium" json:"premium"`
}

type Device struct {
	BotUserID      int64          `db:"bot_user_id" json:"bot_user_id"`
	ApiID          int            `db:"api_id" json:"api_id"`
	ApiHash        string         `db:"api_hash" json:"api_hash"`
	SessionString  string         `db:"session_string" json:"session_string"`
	DeviceModel    string         `db:"device_model" json:"device_model"`
	SystemVersion  string         `db:"system_version" json:"system_version"`
	AppVersion     string         `db:"app_version" json:"app_version"`
	LangPack       string         `db:"lang_pack" json:"lang_pack"`
	LangCode       string         `db:"lang_code" json:"lang_code"`
	SystemLangCode string         `db:"system_lang_code" json:"system_lang_code"`
	Proxy          sql.NullString `db:"proxy" json:"proxy"`
	CreationDate   string         `db:"creation_date" json:"creation_date"`
}
