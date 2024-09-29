package session

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/gotd/td/telegram/dcs"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteSession struct {
	DC        int           `db:"dc_id"`
	Addr      string        `db:"server_address"`
	Port      int           `db:"port"`
	AuthKey   []byte        `db:"auth_key"`
	TakeoutID sql.NullInt32 `db:"takeout_id"`
}

func ConvertSQLiteSessionToTelethonStringSession(
	log *zap.Logger,
	sessionPath string,
) (string, error) {
	db, err := sqlx.Connect("sqlite3", sessionPath)
	if err != nil {
		return "", err
	}

	var session SQLiteSession
	err = db.Get(&session, "SELECT * FROM sessions")
	if err != nil {
		return "", err
	}
	stringSession, err := GetSessionString(session.DC, session.Addr, session.Port, session.AuthKey)
	if err != nil {
		return "", err
	}

	return stringSession, nil
}

func ConvertTDATAToTelethonStringSession(log *zap.Logger, dirname string) (string, error) {
	accounts, err := tdesktop.Read(dirname, nil)
	if err != nil {
		log.Error("Can't read account.", zap.Error(err))
		return "", err
	}
	for _, account := range accounts {
		sd, err := session.TDesktopSession(account)
		if err != nil {
			log.Error("Can't get session data.", zap.Error(err))
			return "", err
		}
		addr := strings.Split(sd.Addr, " ")
		ipAddr := addr[0]
		port, err := strconv.Atoi(addr[1])
		if err != nil {
			return "", err
		}
		stringSession, err := GetSessionString(sd.DC, ipAddr, port, sd.AuthKey)
		return stringSession, err
	}
	return "", errors.New("too many accounts")
}

func GetSessionString(dc int, addr string, port int, authkey []byte) (string, error) {
	// Given parameter should contain version + data
	// where data encoded using pack as '>B4sH256s' or '>B16sH256s'
	// depending on IP type.
	// Some explanation about struct.pack: https://docs.python.org/3/library/struct.html#byte-order-size-and-alignment
	// '>': means big-endian
	// 'B': unsigned char.
	// '4s' or '16s': char[], so bytes. Can be 4 characters long or 16.
	// 'H': unsigned short, so a uint small integer(16bit): uint16.
	// '256s': char[], 256 characters.
	var buf bytes.Buffer
	// | 1    | byte   | DC ID       |
	err := buf.WriteByte(byte(dc))
	if err != nil {
		return "", err
	}

	// | 4/16 | bytes  | IP address  |
	ip := net.ParseIP(addr)
	_, err = buf.Write(ip.To4())
	// | 2    | uint16 | Port        |
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	_, err = buf.Write(portBytes)

	// | 256  | bytes  | Auth key    |
	_, err = buf.Write(authkey[:])

	if err != nil {
		return "", fmt.Errorf("Couldn't pack bytes to buffer: %s", err)
	}

	data := buf.Bytes()
	stringSession := base64.URLEncoding.EncodeToString(data)
	stringSession = fmt.Sprintf("%d%s", 1, stringSession)
	return stringSession, err
}

func EncodeSessionToTelethonString(sessionData *session.Data) (string, error) {
	// Given parameter should contain version + data
	// where data encoded using pack as '>B4sH256s' or '>B16sH256s'
	// depending on IP type.
	// Some explaation about struct.pack: https://docs.python.org/3/library/struct.html#byte-order-size-and-alignment
	// '>': means big-endian
	// 'B': unsigned char.
	// '4s' or '16s': char[], so bytes. Can be 4 characters long or 16.
	// 'H': unsigned short, so a uint small integer(16bit): uint16.
	// '256s': char[], 256 characters.
	var buf bytes.Buffer
	// | 1    | byte   | DC ID       |
	err := buf.WriteByte(byte(sessionData.DC))

	// | 4/16 | bytes  | IP address  |
	var ipAddr string
	var port int
	if sessionData.Addr != "" {
		addr := strings.Split(sessionData.Addr, ":")
		if len(addr) != 2 {
			return "", fmt.Errorf("Invalid datacenter address.")
		}
		ipAddr = addr[0]
		port, _ = strconv.Atoi(addr[1])
	} else {
		list := dcs.Prod()
		for _, option := range list.Options {
			if sessionData.DC == option.ID && !option.Ipv6 {
				ipAddr = option.IPAddress
				port = option.Port
				break
			}
		}
	}
	ip := net.ParseIP(ipAddr)
	_, err = buf.Write(ip.To4())
	// | 2    | uint16 | Port        |
	portBytes := make([]byte, 2)                        // Create a 2-byte array for the port
	binary.BigEndian.PutUint16(portBytes, uint16(port)) // Pack the port into the byte array
	_, err = buf.Write(portBytes)

	// | 256  | bytes  | Auth key    |
	_, err = buf.Write(sessionData.AuthKey[:])

	if err != nil {
		return "", fmt.Errorf("Couldn't pack bytes to buffer: %s", err)
	}

	data := buf.Bytes()
	stringSession := base64.URLEncoding.EncodeToString(data)
	final := fmt.Sprintf("%d%s", 1, stringSession)

	return final, nil
}
