package signature

type Signer interface {
	// Sign - signs the given blob content and returns the signature
	Sign(blobContent []byte) ([]byte, error)
	// VerifySignature - verifies the signature of the given blob content
	VerifySignature(blobContent []byte, signature []byte) error
	// GetPublicKey - returns the public key used for signing
	GetPublicKey() ([]byte, error)
	// ComputeHash - computes the hash of the given blob content
	ComputeHash(blobContent []byte) []byte
}
