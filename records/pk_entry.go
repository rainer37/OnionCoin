package records

type PKEntry struct {
	Key []byte
	Time uint64
}

func New_PKEntry() *PKEntry {
	println("create a new public key entry.")
	return new(PKEntry)
}

func (pke *PKEntry) Get_Key() []byte {
	return pke.Key
}

func (pke *PKEntry) Get_Time() uint64 {
	return pke.Time
}