package repository

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lp/campus-market/internal/util"
)

// fakeSchemaStore is an information_schema stand-in: ExecSQL mutates the fake
// catalog just like MySQL would, so migration steps become deterministic
// without a live database.
type fakeSchemaStore struct {
	columnExists bool
	indexExists  bool

	columnErr error
	indexErr  error
	execErrs  map[string]error // keyed by DDL constant

	executed []string
}

func (f *fakeSchemaStore) ColumnExists(_, _ string) (bool, error) {
	return f.columnExists, f.columnErr
}

func (f *fakeSchemaStore) IndexExists(_, _ string) (bool, error) {
	return f.indexExists, f.indexErr
}

func (f *fakeSchemaStore) ExecSQL(ddl string) error {
	f.executed = append(f.executed, ddl)
	if err := f.execErrs[ddl]; err != nil {
		return err
	}
	switch ddl {
	case tradeOrderAddColumnDDL:
		// MySQL metadata-lock serialisation: a racing loser gets 1060.
		if f.columnExists {
			return &mysql.MySQLError{Number: 1060, Message: "Duplicate column name"}
		}
		f.columnExists = true
	case tradeOrderAddIndexDDL:
		// A racing loser gets 1061; duplicate data would instead yield 1062.
		if f.indexExists {
			return &mysql.MySQLError{Number: 1061, Message: "Duplicate key name"}
		}
		f.indexExists = true
	}
	return nil
}

func TestEnsureTradeOrderActiveKeyMigrations(t *testing.T) {
	tests := []struct {
		name             string
		columnExists     bool
		indexExists      bool
		wantExecutedDDLs []string
		wantColumn       bool
		wantIndex        bool
	}{
		{
			name:             "first migration adds column then index",
			columnExists:     false,
			indexExists:      false,
			wantExecutedDDLs: []string{tradeOrderAddColumnDDL, tradeOrderAddIndexDDL},
			wantColumn:       true,
			wantIndex:        true,
		},
		{
			name:             "column present but index missing repairs only the index",
			columnExists:     true,
			indexExists:      false,
			wantExecutedDDLs: []string{tradeOrderAddIndexDDL},
			wantColumn:       true,
			wantIndex:        true,
		},
		{
			name:             "index present but column missing repairs only the column",
			columnExists:     false,
			indexExists:      true,
			wantExecutedDDLs: []string{tradeOrderAddColumnDDL},
			wantColumn:       true,
			wantIndex:        true,
		},
		{
			name:             "repeat run with both present executes nothing",
			columnExists:     true,
			indexExists:      true,
			wantExecutedDDLs: nil,
			wantColumn:       true,
			wantIndex:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeSchemaStore{columnExists: tt.columnExists, indexExists: tt.indexExists, execErrs: map[string]error{}}
			if err := ensureTradeOrderActiveKey(fake); err != nil {
				t.Fatalf("migration failed: %v", err)
			}
			if len(fake.executed) != len(tt.wantExecutedDDLs) {
				t.Fatalf("executed %v, want %v", fake.executed, tt.wantExecutedDDLs)
			}
			for i, want := range tt.wantExecutedDDLs {
				if fake.executed[i] != want {
					t.Fatalf("ddl #%d = %q, want %q", i, fake.executed[i], want)
				}
			}
			if fake.columnExists != tt.wantColumn || fake.indexExists != tt.wantIndex {
				t.Fatalf("catalog after migration: column=%v index=%v", fake.columnExists, fake.indexExists)
			}
			// The migration may never touch order data: only additive DDL.
			for _, ddl := range fake.executed {
				lowered := strings.ToLower(ddl)
				if strings.Contains(lowered, "update ") || strings.Contains(lowered, "delete ") {
					t.Fatalf("migration must not modify rows: %s", ddl)
				}
			}
		})
	}
}

// A repeated run after a successful migration stays a no-op even when callers
// invoke it many times.
func TestEnsureTradeOrderActiveKeyRepeatedlySafe(t *testing.T) {
	fake := &fakeSchemaStore{execErrs: map[string]error{}}
	for i := 0; i < 5; i++ {
		if err := ensureTradeOrderActiveKey(fake); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}
	// First run adds both pieces; subsequent four runs add nothing.
	if len(fake.executed) != 2 {
		t.Fatalf("expected exactly 2 DDL statements across 5 runs, got %d: %v", len(fake.executed), fake.executed)
	}
}

// Partial state across separate invocations must converge: run 1 adds only
// the column (simulating a crash before the index), run 2 finishes the index.
func TestEnsureTradeOrderActiveKeyConvergesAfterCrash(t *testing.T) {
	fake := &fakeSchemaStore{execErrs: map[string]error{}}

	fake.execErrs[tradeOrderAddIndexDDL] = errors.New("connection lost")
	if err := ensureTradeOrderActiveKey(fake); err == nil {
		t.Fatal("expected error when index DDL fails")
	}
	if len(fake.executed) != 2 || !fake.columnExists || fake.indexExists {
		t.Fatalf("partial migration state wrong: executed=%v column=%v index=%v", fake.executed, fake.columnExists, fake.indexExists)
	}

	// Next boot: column already there, index still missing -> repair index only.
	fake.execErrs = map[string]error{}
	if err := ensureTradeOrderActiveKey(fake); err != nil {
		t.Fatalf("repair run: %v", err)
	}
	if !fake.columnExists || !fake.indexExists {
		t.Fatalf("migration did not converge: column=%v index=%v", fake.columnExists, fake.indexExists)
	}
}

func TestEnsureTradeOrderActiveKeyProbeErrors(t *testing.T) {
	probeErr := errors.New("information_schema unavailable")

	columnFailing := &fakeSchemaStore{columnErr: probeErr, execErrs: map[string]error{}}
	if err := ensureTradeOrderActiveKey(columnFailing); !errors.Is(err, probeErr) {
		t.Fatalf("column probe error: got %v, want %v", err, probeErr)
	}
	if len(columnFailing.executed) != 0 {
		t.Fatalf("no DDL may run after a failed column probe: %v", columnFailing.executed)
	}

	indexFailing := &fakeSchemaStore{columnExists: true, indexErr: probeErr, execErrs: map[string]error{}}
	if err := ensureTradeOrderActiveKey(indexFailing); !errors.Is(err, probeErr) {
		t.Fatalf("index probe error: got %v, want %v", err, probeErr)
	}
	if len(indexFailing.executed) != 0 {
		t.Fatalf("no DDL may run after a failed index probe: %v", indexFailing.executed)
	}
}

// Pre-existing duplicate active orders make the unique index fail with 1062;
// the migration must surface it instead of reporting success.
func TestEnsureTradeOrderActiveKeyDuplicateDataFailsLoudly(t *testing.T) {
	dupErr := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}
	fake := &fakeSchemaStore{
		columnExists: true,
		indexExists:  false,
		execErrs:     map[string]error{tradeOrderAddIndexDDL: dupErr},
	}
	err := ensureTradeOrderActiveKey(fake)
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		t.Fatalf("expected 1062 failure on dirty data, got %v", err)
	}
	if fake.indexExists {
		t.Fatal("unique index must not be reported present after a failed CREATE")
	}
}

func TestMapTradeOrderCreateError(t *testing.T) {
	duplicate := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}
	fkErr := &mysql.MySQLError{Number: 1452, Message: "foreign key"}
	plain := errors.New("boom")

	if got := mapTradeOrderCreateError(duplicate); !errors.Is(got, util.ErrConflict) {
		t.Fatalf("1062 -> %v, want ErrConflict", got)
	}
	if got := mapTradeOrderCreateError(fkErr); got != fkErr {
		t.Fatalf("1452 -> %v, want the original error", got)
	}
	if got := mapTradeOrderCreateError(plain); got != plain {
		t.Fatalf("plain error -> %v, want the original error", got)
	}
	if got := mapTradeOrderCreateError(nil); got != nil {
		t.Fatalf("nil -> %v, want nil", got)
	}
}

// racingCatalog models several application instances sharing one MySQL
// catalog. Its mutex plays the role of the table metadata lock that
// serialises DDL: the first ADD wins and every later ADD fails with
// 1060/1061, exactly like MySQL reports to the losing instance.
type racingCatalog struct {
	mu                             sync.Mutex
	columnExists                   bool
	indexExists                    bool
	indexDDLErr                    *mysql.MySQLError // persistent failure (e.g. 1062 dirty data)
	createdColumns, createdIndexes int
}

func (c *racingCatalog) columnPresent() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.columnExists, nil
}

func (c *racingCatalog) indexPresent() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.indexExists, nil
}

func (c *racingCatalog) exec(ddl string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch ddl {
	case tradeOrderAddColumnDDL:
		if c.columnExists {
			return &mysql.MySQLError{Number: 1060, Message: "Duplicate column name"}
		}
		c.columnExists = true
		c.createdColumns++
	case tradeOrderAddIndexDDL:
		if c.indexDDLErr != nil {
			return c.indexDDLErr
		}
		if c.indexExists {
			return &mysql.MySQLError{Number: 1061, Message: "Duplicate key name"}
		}
		c.indexExists = true
		c.createdIndexes++
	}
	return nil
}

// racingInstance is one application instance; it only forwards to the shared
// catalog (and records the DDL it attempted, guarded by its own lock).
type racingInstance struct {
	shared   *racingCatalog
	mu       sync.Mutex
	attempts []string
}

func (i *racingInstance) ColumnExists(string, string) (bool, error) { return i.shared.columnPresent() }
func (i *racingInstance) IndexExists(string, string) (bool, error)  { return i.shared.indexPresent() }
func (i *racingInstance) ExecSQL(ddl string) error {
	i.mu.Lock()
	i.attempts = append(i.attempts, ddl)
	i.mu.Unlock()
	return i.shared.exec(ddl)
}

func racingInstances(n int, shared *racingCatalog) []schemaStore {
	stores := make([]schemaStore, n)
	for i := range stores {
		stores[i] = &racingInstance{shared: shared}
	}
	return stores
}

func runEnsureConcurrently(t *testing.T, stores []schemaStore) []error {
	t.Helper()
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make([]error, len(stores))
	wg.Add(len(stores))
	for i, st := range stores {
		go func(i int, st schemaStore) {
			defer wg.Done()
			<-start
			errs[i] = ensureTradeOrderActiveKey(st)
		}(i, st)
	}
	close(start)
	wg.Wait()
	return errs
}

// Multiple instances booting on an empty table: the column and the unique
// index are each installed exactly once, and no instance exits because
// another instance won the creation race (1060/1061 losers are tolerated).
func TestEnsureTradeOrderActiveKeyConcurrentFirstBoot(t *testing.T) {
	const instances = 8
	shared := &racingCatalog{}
	stores := racingInstances(instances, shared)

	errs := runEnsureConcurrently(t, stores)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("instance %d aborted after losing a creation race: %v", i, err)
		}
	}
	if shared.createdColumns != 1 || shared.createdIndexes != 1 {
		t.Fatalf("column created %d times, index %d times, want 1/1",
			shared.createdColumns, shared.createdIndexes)
	}
	if !shared.columnExists || !shared.indexExists {
		t.Fatalf("catalog not converged: column=%v index=%v", shared.columnExists, shared.indexExists)
	}

	// Repeat startup with both objects present: every instance is a no-op.
	errs = runEnsureConcurrently(t, stores)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("instance %d failed on repeat boot: %v", i, err)
		}
	}
	if shared.createdColumns != 1 || shared.createdIndexes != 1 {
		t.Fatalf("repeat boot re-created schema objects: columns=%d indexes=%d",
			shared.createdColumns, shared.createdIndexes)
	}
}

// The column is already installed; instances only race the remaining index.
// It is created exactly once and the 1061 losers still start successfully.
func TestEnsureTradeOrderActiveKeyConcurrentIndexOnlyRepair(t *testing.T) {
	const instances = 8
	shared := &racingCatalog{columnExists: true}
	stores := racingInstances(instances, shared)

	errs := runEnsureConcurrently(t, stores)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("instance %d aborted on a 1061 index race: %v", i, err)
		}
	}
	if shared.createdColumns != 0 || shared.createdIndexes != 1 {
		t.Fatalf("columns created=%d want 0; indexes created=%d want 1",
			shared.createdColumns, shared.createdIndexes)
	}
}

// Failure recovery under concurrency: duplicate active orders make the unique
// index fail with 1062, so every instance must surface the failure and the
// guard stays absent; after the data is cleaned, the next concurrent boot
// installs the index exactly once and all instances succeed.
func TestEnsureTradeOrderActiveKeyConcurrentFailureRecovery(t *testing.T) {
	const instances = 4
	shared := &racingCatalog{
		columnExists: true,
		indexDDLErr:  &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
	}
	stores := racingInstances(instances, shared)

	errs := runEnsureConcurrently(t, stores)
	for i, err := range errs {
		var mysqlErr *mysql.MySQLError
		if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
			t.Fatalf("instance %d: got %v, want 1062 fatal on dirty data", i, err)
		}
	}
	if shared.indexExists || shared.createdIndexes != 0 {
		t.Fatalf("index must not exist after failed attempts: exists=%v created=%d",
			shared.indexExists, shared.createdIndexes)
	}

	// Data cleaned: the index DDL now succeeds; the concurrent restart converges.
	shared.indexDDLErr = nil
	errs = runEnsureConcurrently(t, stores)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("instance %d after recovery: %v", i, err)
		}
	}
	if !shared.indexExists || shared.createdIndexes != 1 {
		t.Fatalf("after recovery index exists=%v created=%d, want true/1",
			shared.indexExists, shared.createdIndexes)
	}
}

func TestIsAlreadyExistsDDLError(t *testing.T) {
	wrapped := fmt.Errorf("alter table: %w", &mysql.MySQLError{Number: 1061})
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "1060 duplicate column", err: &mysql.MySQLError{Number: 1060}, want: true},
		{name: "1061 duplicate key", err: &mysql.MySQLError{Number: 1061}, want: true},
		{name: "1062 duplicate value stays fatal", err: &mysql.MySQLError{Number: 1062}, want: false},
		{name: "1452 other mysql error stays fatal", err: &mysql.MySQLError{Number: 1452}, want: false},
		{name: "wrapped 1061 still recognised", err: wrapped, want: true},
		{name: "non mysql error", err: errors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyExistsDDLError(tt.err); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
