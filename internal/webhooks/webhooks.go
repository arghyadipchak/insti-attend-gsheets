package webhooks

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/attendances"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/collections"
)

var errBadToken = errors.New("missing or malformed bearer token")

func Register(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/wbh/{slug}", handler)
		return se.Next()
	})
}

func handler(re *core.RequestEvent) error {
	if ct := re.Request.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return re.BadRequestError("invalid content type, expected application/json", nil)
	}

	slug := re.Request.PathValue("slug")

	webhook, findErr := collections.FindWebhookBySlug(re.App, slug)
	if findErr != nil || !webhook.Enabled() {
		return re.NotFoundError("webhook not found", findErr)
	}

	secret, err := bearerToken(re.Request)
	if err != nil {
		return re.UnauthorizedError(err.Error(), nil)
	}

	authenticated, authErr := webhook.Authenticate(re.App, secret)
	if authErr != nil {
		return re.InternalServerError("authentication failed", authErr)
	}
	if !authenticated {
		return re.UnauthorizedError("invalid bearer token", nil)
	}

	records, err := attendances.ReadAttendanceFromIO(re.Request.Body)
	if err != nil {
		return re.BadRequestError("invalid attendance records: "+err.Error(), nil)
	}
	if len(records) == 0 {
		return re.NoContent(http.StatusNoContent)
	}

	for _, record := range records {
		record.Synced = false
	}

	spreadsheetIds := webhook.SpreadsheetIDs()
	if len(spreadsheetIds) == 0 {
		return re.NoContent(http.StatusNoContent)
	}

	id, createErr := collections.CreateAttendanceRecord(re.App, records, spreadsheetIds)
	if createErr != nil {
		return re.InternalServerError("failed to record attendance", createErr)
	}

	return re.JSON(http.StatusCreated, map[string]string{"id": id})
}

func bearerToken(req *http.Request) (string, error) {
	header := strings.TrimSpace(req.Header.Get("Authorization"))
	if header == "" {
		return "", errBadToken
	}

	const prefix = "Bearer "

	token, found := strings.CutPrefix(header, prefix)
	if !found {
		return "", errBadToken
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", errBadToken
	}

	return token, nil
}
