package auth_test

import (
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/golang-jwt/jwt/v4"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to authatnicate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privatePem))
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", success, testID)

			const keyID = "7a7fb378-d885-43ad-aa25-a0b33bca287f"
			lookup := func(kid string) (*rsa.PublicKey, error) {
				switch kid {
				case keyID:
					return &privateKey.PublicKey, nil
				}
				return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
			}

			a, err := auth.New("RS256", lookup, auth.Keys{keyID: privateKey})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", success, testID)

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "financify service",
					Subject:   "9920bcc5-c203-417a-8706-5f007b0357cc",
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := a.GenerateToken(keyID, claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t\tTest %d:\texp: %d", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %d", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected number of roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected number of roles.", success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected roles: %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected roles.", success, testID)
		}
	}
}

var privatePem = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAteN1TlVJLKlTrTVEqAWeop1v14dL20wy6LdPrpE1RPb23Xek
v/EIDZlOp0MYvxEnXCo7Dh2cKoZPGWbnS4OhtPhWxRn0O34auJ5e0polWI8EZiRZ
R3ZOSO6+VkDf0C44DRvyg5Nusgb8fa97YKoiSIXRqmlEWtSOBtrvHJgEW3GI8Skn
G+LvcEvwFr2m45UdJVWtDuPD7rQNCu1yWRu8qE9OFlm7lI10mXpMnEvtPx8KxOsA
s5GbmonCJZRPVWTN7aowp1GuYgeomfqBfBIjPFTLUJLTMaRlh2i53JaQp606oL9A
MoDUZpl0NcAarFllCr4kvcVD74L0Mozn6Qi7nwIDAQABAoIBAElOM9/vNX+fes7r
EhGZujaVtxapO6RVkIsEHkQf19VEp3fYmXiWPwWkDPQSca3HzxIxHv3wZxkoaka1
l3Byy8Bw+h+T9z/m8gQIJ/U/FOAdO8uiyKypfKGePu3qVYnEpuh5pALtb4amlCpf
iB0MVKbf8AF7TYZB9j/DCu1+QvtLFxPXe2sM3/ksqWrkHEPa/aYgU0NejhJe8Oeo
qqW+zIk4w4G46oFTxDeHBKxvTwPehmv8qjKWTseYL4Hwa5p4kvCDYrnBl5OSymC+
vZQdYopftUs1PX/iEO8YcICdy40NjbOrV0xI+ZHeq5m53Cr3xEY7n0o00HoHg61q
RZhRLykCgYEAxK5zUaE9oCRgmeAp96LFDhOLtHuuaZyPQ+nqJoXpfLt0PwE4gv9E
/E80WBZdJNP0/h3XYa8qCePyTRpnSdE/NJY/5U1XAsZAPlaRbb2qHudj6Cg9iD5V
Iszwwza3QbFh/aYLKGt1coQg+Ja8RuOQIzuUBWhSId/nfEw/uLf5QZMCgYEA7L7c
pgzNHYlNsnfFbGWQJMTpIyfjcxDWaXEYDcz1upliMylwxbpuPqrLqrF9WcpHOtS8
Wc/+dCml5cg2kqK5WkENWoI6UNboUhEADKGQDI2ZYTSUNSlCtgZzqf3nVPt/YnT1
38eVpxQpAxNRHSfqa+IQuUlLclLVHFaOmM1DFUUCgYBFvrcWE1+PEldPObaoIghO
3Y+FCPbobKRBKQnnb0VE/hRS41Pu4CbOcifVtNiC3sbZ9isScNMvfq3Fub825gTL
2Rv/bFWWnkbZ1Ejt6XwSSWucP+jSD4iRNquKDjUeDpD5KZB7XN/hJAmtHYbWfIv4
coAjCsNVT9j+sutFzbeOEQKBgQC5SXevjf1KzJc+wnaFK8fwvxwoI6Pj/p2Q0K6e
vnbjoAA3Qou8dPirm0jjQx50E9hDtxPixuLDT4VDnbr4cNrYRGmLGLlDY69X6246
dIglCv2sElacdLp9c/c6aDmRTXSZPijhB3ec2C5w9cFaLE9QOBIWscKWqzWXhDb+
aEfEcQKBgHNW2oJu0ovsfl959BpBYGKVZ7RrDB/h/BEr7gEbibbA6DjLCXNfDWra
BBXP3xWqyd3XbfpdKnElDCp5USuSnvTZDfmREAmmgw0ar9f5fyG0VnlriipiNrm2
G0jHHQxn/A3NstR1z0yg4zGJjo73+Cpf58y2qIaor5axIaL8sqhs
-----END RSA PRIVATE KEY-----`
