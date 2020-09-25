package method

import (
	"errors"
	"fmt"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/meshplus/bitxid/internal/common/registry"
	"github.com/sirupsen/logrus"
)

// the rule of status code:
// end with 1 (001, 101, 301, etc.) means on audit
// end with 5 (005, 105, 205, 305, etc.) means audit failed
// end with 0 (010, 110, 200, 310, etc.) means good
// difference between
const (
	ApplyAudit      int = 001
	ApplyFailed     int = 005
	ApplySuccess    int = 010
	RegisterAudit   int = 101
	RegisterFailed  int = 105
	RegisterSuccess int = 110
	Normal          int = 200
	Frozen          int = 205
	UpdateAudit     int = 301
	UpdateFailed    int = 305
	UpdateSuccess   int = 310
)

// Registry .
type Registry struct {
	table   *registry.Table
	logger  logrus.FieldLogger
	network int
}

// Item is item in Registry
// registry table is used together with docdb
// we suggest to store large data off-chain(in docdb)
// and only some frequently used data on-chain(in cache)
type Item struct {
	key     string // primary key of the item, like a did
	docAddr string // addr where the doc file stored
	docHash string // hash of the doc file
	status  int    // status of the item
	cache   []byte // onchain storage part
}

type did string

// New a MethodRegistry
func New(S storage.Storage, L logrus.FieldLogger) (*Registry, error) {
	rt, err := registry.NewTable(S)
	if err != nil {
		fmt.Println("[registry.NewTable] err", err)
		return nil, err
	}
	return &Registry{
		table:   rt,
		logger:  L,
		network: 1,
	}, nil
}

// SetupGenesis set up genesis to boot the whole methed system
func SetupGenesis() {}

func Start() {}

// Apply apply rights for a new methd
func (R *Registry) Apply(caller did, method string, sig []byte) error {
	// check if did exists
	// ..

	// check if Method Name meets standard
	// ..

	// check if Method exists
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == true {
		return errors.New("The Method is ALREADY existed")
	}
	// creates item in table
	err = R.table.CreateItem([]byte(method),
		Item{
			key:    method,
			status: ApplyAudit,
		})
	if err != nil {
		R.logger.Error("[R.table.CreateItem] err:", err)
		return err
	}
	return nil
}

// AuditApply .
func (R *Registry) AuditApply(caller did, method string, result bool, sig []byte) error {
	if !R.isStatus(method, ApplyAudit) {
		return errors.New("Can not be AuditApply for current status")
	}
	if result {
		R.audit(method, ApplySuccess)
	}
	R.audit(method, ApplyFailed)
	return nil
}

func (R *Registry) isStatus(method string, status int) bool {
	item := Item{}
	R.table.GetItem([]byte(method), &item)
	if item.status == status {
		return true
	}
	return false
}

// Audit .
func (R *Registry) audit(method string, status int) error {
	exist, err := R.HasMethod(method)
	if err != nil {
		R.logger.Error("[R.HasMethod] err:", err)
		return err
	}
	if exist == false {
		return errors.New("The Method NOT existed")
	}
	err = R.table.UpdateItem([]byte(method),
		Item{
			key:    method,
			status: status,
		})
	if err != nil {
		R.logger.Error("[R.table.UpdateItem] err:", err)
		return err
	}
	return nil
}

// HasMethod .
func (R *Registry) HasMethod(method string) (bool, error) {
	exist, err := R.table.HasItem([]byte(method))
	if err != nil {
		R.logger.Error("[R.table.HasItem] err:", err)
		return false, err
	}
	return exist, err
}
