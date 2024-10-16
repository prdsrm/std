package postgres

import (
	"github.com/jmoiron/sqlx"
)

func InsertNewBot(
	db *sqlx.DB,
	phone string,
	username string,
	password string,
	userID int64,
	title string,
) error {
	_, err := db.Exec(`
            INSERT INTO
              bots (
                phone_number,
                username,
                password,
                user_id,
				title
              )
            VALUES
              ($1, $2, $3, $4, $5);
		`, phone, username, password, userID, title)
	if err != nil {
		return err
	}
	return nil
}

func DeleteBotByUserID(db *sqlx.DB, botUserID int64) error {
	_, err := db.Exec("DELETE FROM bots WHERE user_id=$1", botUserID)
	if err != nil {
		return err
	}
	return err
}

func GetAllBots(db *sqlx.DB) ([]Bot, error) {
	var bots []Bot
	query := `SELECT * FROM bots`
	err := db.Select(&bots, query)
	if err != nil {
		return nil, err
	}
	return bots, nil
}

func GetRandomBot(db *sqlx.DB) (*Bot, error) {
	bot := Bot{}
	query := `SELECT * FROM bots ORDER BY RANDOM() LIMIT 1`
	err := db.Get(&bot, query)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func GetBotByUserID(db *sqlx.DB, botUserID int64) (*Bot, error) {
	bot := Bot{}
	err := db.Get(&bot, "SELECT * FROM bots WHERE user_id=$1", botUserID)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func GetRandomDevice(db *sqlx.DB, botUserID int64) (*Device, error) {
	device := Device{}
	query := `SELECT * FROM devices WHERE bot_user_id=$1 ORDER BY RANDOM() LIMIT 1`
	err := db.Get(&device, query, botUserID)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func InsertNewDevice(db *sqlx.DB, userID int64, apiID int, apiHash string, sessionString string, deviceModel string, systemVersion string, appVersion string, langPack string, langCode string, systemLangCode string, proxy string) error {
	_, err := db.Exec(`
            INSERT INTO
              devices (
                bot_user_id,
				api_id,
                api_hash,
                session_string,
                device_model,
				system_version,
				app_version,
				lang_pack,
				lang_code,
				system_lang_code,
				proxy
              )
            VALUES
              ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
		`, userID, apiID, apiHash, sessionString, deviceModel, systemVersion, appVersion, langPack, langCode, systemLangCode, proxy)
	if err != nil {
		return err
	}
	return nil
}

func DeleteDeviceBySessionString(db *sqlx.DB, session string) error {
	_, err := db.Exec("DELETE FROM bots WHERE session_string=$1", session)
	if err != nil {
		return err
	}
	return err
}
