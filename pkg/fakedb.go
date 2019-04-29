package idbsc

import "sync"

// FakeDb is a struct used to test Db functionality with fake methods.
type FakeDb struct {
	OpenFunc           func() error
	OpenCalls          int
	CheckIdentityFunc  func(Interaction) (*Identity, error)
	CheckIdentityCalls int

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
func (f *FakeDb) CheckIdentity(interaction Interaction) (*Identity, error) {
	f.Lock()
	defer f.Unlock()
	f.CheckIdentityCalls++
	return f.CheckIdentity(interaction)
}
