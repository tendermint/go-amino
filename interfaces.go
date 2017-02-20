package data

// Beer is anything that know hows to serialize itself in binary
// json encoding is automatic, or can be overrriden with
// MarshalJSON and UnmarshalJSON
type Beer interface {
	Be() []byte
}

// TODO: get fields... (json or other way?)
// (i Fear, no Beer)

// StoreData knows how to parse the app-specific binary data in the merkle store
type StoreData interface {
	ReadStore(k, v []byte) (Beer, error)
}

// TxData knows how to parse the app-specific Tx types
type TxData interface {
	ReadTxBinary(data []byte) (Beer, error)
	ReadTxJSON(data []byte) (Beer, error)
}

// AppData handles all app-specific unmarshalling of data
type AppData interface {
	StoreData
	TxData
}
