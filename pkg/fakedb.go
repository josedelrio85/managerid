package passport

import "sync"

// FakeDb is a struct used to test Db functionality with fake methods.
type FakeDb struct {
	OpenFunc         func() error
	OpenCalls        int
	GetIdentityFunc  func(Interaction) (*Identity, error)
	GetIdentityCalls int
	CloseFunc        func() error
	CloseCalls       int
	CreateTableFunc  func() error
	CreateTableCalls int

	sync.Mutex
}

// Open is a method to test Open function
func (f *FakeDb) Open() error {
	f.Lock()
	defer f.Unlock()
	f.OpenCalls++
	return f.Open()
}

// GetIdentity is a method to test GetIdentity function
func (f *FakeDb) GetIdentity(interaction Interaction) (*Identity, error) {
	f.Lock()
	defer f.Unlock()
	f.GetIdentityCalls++
	return f.GetIdentityFunc(interaction)
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
