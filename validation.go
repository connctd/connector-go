package connector

import "github.com/go-fed/httpsig"

var (
	// SignatureAlgorithm defines which algorithm is used for generating the signature
	SignatureAlgorithm = httpsig.ED25519

	// SignedHeaders defines which headers are signed together with the message body
	SignedHeaders = []string{httpsig.RequestTarget, "date", "Digest"}

	// HashAlgorithm defines which algorithm is used to calculate the body digest
	HashAlgorithm = httpsig.DigestSha256

	// KeyIDPublication identifies the public key created during publication of connector
	KeyIDPublication = KeyID("PublicationPublicKey")
)

// KeyID indicates which key was/has to be used for a cryptographic operation
type KeyID string
