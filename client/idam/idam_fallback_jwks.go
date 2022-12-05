package idam

var (
	// IDAMFallbackJWKMap contains IDAM JWKs to be used when IDAM is not available at startup
	IDAMFallbackJWKMap = JWKMap{
		"token-signing-keypair": JWK{
			KTY: "RSA",
			E:   "AQAB",
			KID: "token-signing-keypair",
			//nolint
			N: "xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw",
		},
	}
)
