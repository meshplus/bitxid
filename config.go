package bitxid

func getAdminDoc() *AccountDoc {
	doc := &AccountDoc{}
	doc.ID = "did:bitxhub:appchain001:0x00000001"
	doc.Type = int(ChainDIDType)
	pk := PubKey{
		ID:           "KEY#1",
		Type:         "Ed25519",
		PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
	}
	doc.PublicKey = []PubKey{pk}
	auth := Auth{
		PublicKey: []string{"KEY#1"},
	}
	doc.Authentication = []Auth{auth}
	return doc
}

func genesisAccountDoc() *AccountDoc {
	return &AccountDoc{
		BasicDoc: BasicDoc{
			ID:   "did:bitxhub:appchain001:0x00000001",
			Type: int(AccountDIDType),
			PublicKey: []PubKey{
				{
					ID:           "KEY#1",
					Type:         "Ed25519",
					PublicKeyPem: "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV",
				},
			},
			Authentication: []Auth{
				{PublicKey: []string{"KEY#1"}},
			},
		},
	}
}

func genesisChainDoc() *ChainDoc {
	return &ChainDoc{
		BasicDoc: BasicDoc{
			ID:   "did:bitxhub:relayroot:.",
			Type: int(ChainDIDType),
			PublicKey: []PubKey{
				{
					ID:           "KEY#1",
					Type:         "Secp256k1",
					PublicKeyPem: "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71",
				},
			},
			Controller: DID("did:bitxhub:relayroot:0x00000001"),
			Authentication: []Auth{
				{PublicKey: []string{"KEY#1"}},
			},
		},
	}
}
