package domain

import "time"

// WalletInfo is the projection of a linked wallet identity returned to
// authenticated callers via ListWallets.
type WalletInfo struct {
	Address  string    `bson:"external_id" json:"address"`
	ChainId  int64     `bson:"-" json:"chain_id"`
	LinkedAt time.Time `bson:"linked_at" json:"linked_at"`
}
