package data

// Byter is anything that know hows to serialize itself in binary
// json encoding is automatic, or can be overrriden with
// MarshalJSON and UnmarshalJSON
type Byter interface {
	Bytes() []byte
}

// StoreData knows how to parse the app-specific binary data in the merkle store
type StoreData interface {
	ReadStore(k, v []byte) (Byter, error)
}

// TxData knows how to parse the app-specific Tx types
type TxData interface {
	ReadTxBinary(data []byte) (Byter, error)
	ReadTxJSON(data []byte) (Byter, error)
}

// AppData handles all app-specific unmarshalling of data
type AppData interface {
	StoreData
	TxData
}
