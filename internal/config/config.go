package config

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/util"
)

var (
	SpreadsheetId string

	ColRollIndex, _  = util.ColumnLetterToIndex("A")
	ColStartIndex, _ = util.ColumnLetterToIndex("B")
	ColDateFormat    = "2 Jan"

	RowHeader uint32 = 1
	RowStart  uint32 = 2
	RowFormat        = "P"
	RowIsTime        = false

	CredentialsFile = "credentials.json"
	WebhookAddr     = ":8080"
	AuthToken       string
)

func Init() {
	if value, exists := os.LookupEnv("SPREADSHEET_ID"); exists {
		SpreadsheetId = value
	} else {
		log.Fatalln("env variable not set: SPREADSHEET_ID")
	}

	if value, exists := os.LookupEnv("CREDENTIALS_FILE"); exists {
		CredentialsFile = value
	}
	fileInfo, err := os.Stat(CredentialsFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("credentials file error:", CredentialsFile, "does not exist")
		} else if os.IsPermission(err) {
			log.Fatalln("credentials file error:", CredentialsFile, "permission denied")
		} else {
			log.Fatalln("credentials file error:", CredentialsFile, err)
		}
	}
	if fileInfo.IsDir() {
		log.Fatalln("credentials file error:", CredentialsFile, "is a directory")
	}

	if value, exists := os.LookupEnv("COL_ROLL"); exists {
		if ColRollIndex, err = util.ColumnLetterToIndex(value); err != nil {
			log.Fatalln("column COL_ROLL error:", value, err)
		}
	}

	if value, exists := os.LookupEnv("COL_START"); exists {
		if ColStartIndex, err = util.ColumnLetterToIndex(value); err != nil {
			log.Fatalln("column COL_START error:", value, err)
		}
	}

	if value, exists := os.LookupEnv("COL_DATE_FORMAT"); exists {
		if _, err := time.Parse(value, "2 Jan"); err != nil {
			log.Fatalln("format COL_DATE_FORMAT invalid:", value, err)
		}
		ColDateFormat = value
	}

	if value, exists := os.LookupEnv("ROW_HEADER"); exists {
		value, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			log.Fatalln("row ROW_HEADER error:", value, err)
		}
		if value < 1 {
			log.Fatalln("row ROW_HEADER error:", value, "must be >= 1")
		}
		RowHeader = uint32(value)
	}

	if value, exists := os.LookupEnv("ROW_START"); exists {
		value, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			log.Fatalln("row ROW_START error:", value, err)
		}
		if value < 1 {
			log.Fatalln("row ROW_START error:", value, "must be >= 1")
		}
		RowStart = uint32(value)
	}

	if value, exists := os.LookupEnv("ROW_FORMAT"); exists {
		RowFormat = value
		if _, err := time.Parse(value, "15:04:05"); err == nil {
			RowIsTime = true
		}
	}

	if value, exists := os.LookupEnv("WEBHOOK_ADDR"); exists {
		if _, err := net.ResolveTCPAddr("tcp", value); err != nil {
			log.Fatalln("addr WEBHOOK_ADDR invalid:", value, err)
		}
		WebhookAddr = value
	}

	if value, exists := os.LookupEnv("AUTH_TOKEN"); exists {
		AuthToken = value
	}
}
