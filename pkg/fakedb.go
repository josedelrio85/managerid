package idbsc

import "sync"

// FakeDb is a struct used to test Db functionality with fake methods.
type FakeDb struct {
	OpenFunc           func() error
	OpenCalls          int
	CheckIdentityFunc  func(Interaction) (*Identitygorm, error)
	CheckIdentityCalls int
	CloseFunc          func() error
	CloseCalls         int
	CreateTableFunc    func() error
	CreateTableCalls   int

	sync.Mutex
}

// Open is a method to test Open function
func (f *FakeDb) Open() error {
	f.Lock()
	defer f.Unlock()
	f.OpenCalls++
	return f.Open()
}

// CheckIdentity is a method to test CheckIdentity function
func (f *FakeDb) CheckIdentity(interaction Interaction) (*Identitygorm, error) {
	f.Lock()
	defer f.Unlock()
	f.CheckIdentityCalls++
	return f.CheckIdentityFunc(interaction)
}

// Close is a method to test Close function
func (f *FakeDb) Close() {
	f.Lock()
	defer f.Unlock()
	f.CloseCalls++
	f.CloseFunc()
}

// CreateTable is a method to test CreateTable function
func (f *FakeDb) CreateTable() error {
	f.Lock()
	defer f.Unlock()
	f.CreateTableCalls++
	return f.CreateTableFunc()
}
