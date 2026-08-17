package collections

import (
	"crypto/subtle"
	"errors"

	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*Webhook)(nil)

type Webhook struct{ core.BaseRecordProxy }

func WebhookFromRecord(r *core.Record) *Webhook {
	return &Webhook{core.BaseRecordProxy{Record: r}}
}

func (w *Webhook) Record() *core.Record { return w.BaseRecordProxy.Record }

func (w *Webhook) Slug() string        { return w.GetString("slug") }
func (w *Webhook) SetSlug(slug string) { w.Set("slug", slug) }

func (w *Webhook) Name() string        { return w.GetString("name") }
func (w *Webhook) SetName(name string) { w.Set("name", name) }

func (w *Webhook) Enabled() bool     { return w.GetBool("enabled") }
func (w *Webhook) SetEnabled(v bool) { w.Set("enabled", v) }

func (w *Webhook) SpreadsheetIDs() []string { return w.GetStringSlice(Spreadsheets) }

func (w *Webhook) Authenticate(app core.App, secret string) (bool, *e.Error) {
	if err := expandOne(app, w, Tokens); err != nil {
		return false, err
	}

	secretBytes := []byte(secret)
	for _, tokenRecord := range w.ExpandedAll(Tokens) {
		tokenSecret := []byte(TokenFromRecord(tokenRecord).Secret())
		if subtle.ConstantTimeCompare(tokenSecret, secretBytes) == 1 {
			return true, nil
		}
	}

	return false, nil
}

func FindWebhookBySlug(app core.App, slug string) (*Webhook, error) {
	record, err := app.FindFirstRecordByData(Webhooks, "slug", slug)
	if err != nil {
		return nil, e.NewError("failed to query webhook by slug", err).AddContext("slug", slug)
	}
	if record == nil {
		err := errors.New("not found")
		return nil, e.NewError("webhook not found for slug", err).AddContext("slug", slug)
	}
	return WebhookFromRecord(record), nil
}
