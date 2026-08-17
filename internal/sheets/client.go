package sheets

import (
	"context"
	"sync"

	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"

	"google.golang.org/api/option"
	gsheets "google.golang.org/api/sheets/v4"
)

type Client struct {
	*gsheets.Service
	sheetIDCache sync.Map
}

func NewClient(credentialsJSON string) (*Client, *e.Error) {
	authOption := option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(credentialsJSON))
	service, err := gsheets.NewService(context.Background(), authOption)
	if err != nil {
		return nil, e.NewError("failed to create Google Sheets service client", err)
	}
	return &Client{Service: service}, nil
}

func (c *Client) FetchTitle(spreadsheetId string) (string, *e.Error) {
	resp, err := c.Spreadsheets.Get(spreadsheetId).Fields("properties.title").Do()
	if err != nil {
		return "", e.NewError("failed to fetch spreadsheet title", err)
	}
	return resp.Properties.Title, nil
}
