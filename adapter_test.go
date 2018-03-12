package boltadapter

import (
	"log"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/util"
)

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes := e.GetPolicy()
	log.Print("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func TestAdapter(t *testing.T) {

	// Because the DB is empty at first,
	// so we need to load the policy from the file adapter (.CSV) first.
	e := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	db, err := bolt.Open("testdata/db.dat", 0600, nil)
	if err != nil {
		t.Fatalf("error opening db: %s\n", err)
	}

	defer func() {
		db.Close()
		if _, err := os.Stat("testdata/db.dat"); err == nil {
			os.Remove("testdata/db.dat")
		}
	}()

	a := NewAdapter(db)
	// This is a trick to save the current policy to the DB.
	// We can't call e.SavePolicy() because the adapter in the enforcer is still the file adapter.
	// The current policy means the policy in the Casbin enforcer (aka in memory).
	a.SavePolicy(e.GetModel())

	// Clear the current policy.
	e.ClearPolicy()
	testGetPolicy(t, e, [][]string{})

	// Load the policy from DB.
	a.LoadPolicy(e.GetModel())
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	a = NewAdapter(db)
	e = casbin.NewEnforcer("examples/rbac_model.conf", a)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

}
