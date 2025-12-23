// Package factory provides factory functions to create instances of cryptographic services.
package factory

import (
	kbxCrp "github.com/kubex-ecosystem/kbx/tools/security/crypto"
	kbxKrs "github.com/kubex-ecosystem/kbx/tools/security/external"
	kbxSec "github.com/kubex-ecosystem/kbx/tools/security/interfaces"
)

type CryptoService = kbxSec.ICryptoService
type KeyringService = kbxSec.IKeyringService

func NewCryptoService() CryptoService {
	return kbxCrp.NewCryptoService()
}

func NewKeyringService(keyringServiceName, keyringServicePath string) KeyringService {
	return kbxKrs.NewFileKeyringService(keyringServiceName, keyringServicePath)
}
