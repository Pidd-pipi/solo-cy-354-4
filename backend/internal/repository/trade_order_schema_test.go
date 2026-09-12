package repository

import (
	"errors"
	"strings"
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
		f.columnExists = true
	case tradeOrderAddIndexDDL:
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
