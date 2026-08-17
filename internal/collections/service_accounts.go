package collections

import (
	"github.com/pocketbase/pocketbase/core"

	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/sheets"
)

var _ core.RecordProxy = (*ServiceAccount)(nil)

type ServiceAccount struct{ core.BaseRecordProxy }

func ServiceAccountFromRecord(r *core.Record) *ServiceAccount {
	return &ServiceAccount{core.BaseRecordProxy{Record: r}}
}

func (sa *ServiceAccount) Record() *core.Record { return sa.BaseRecordProxy.Record }

func (sa *ServiceAccount) Name() string            { return sa.GetString("name") }
func (sa *ServiceAccount) CredentialsJSON() string { return sa.GetString("credentials") }

func (sa *ServiceAccount) NewClient() (*sheets.Client, *e.Error) {
	client, err := sheets.NewClient(sa.CredentialsJSON())
	if err != nil {
		return nil, err
	}
	return client, nil
}

func registerServiceAccountHooks(app core.App) error {
	app.OnRecordValidate(ServiceAccounts).BindFunc(validateServiceAccount)
	return nil
}

func validateServiceAccount(re *core.RecordEvent) error {
	sa := ServiceAccountFromRecord(re.Record)
	if _, err := sa.NewClient(); err != nil {
		return e.NewError("failed to validate service account", err)
	}
	return re.Next()
}
