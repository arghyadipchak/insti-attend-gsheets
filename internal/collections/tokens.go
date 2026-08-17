package collections

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*Token)(nil)

type Token struct{ core.BaseRecordProxy }

func TokenFromRecord(r *core.Record) *Token {
	return &Token{core.BaseRecordProxy{Record: r}}
}

func (t *Token) Record() *core.Record { return t.BaseRecordProxy.Record }

func (t *Token) Name() string   { return t.GetString("name") }
func (t *Token) Secret() string { return t.GetString("secret") }
