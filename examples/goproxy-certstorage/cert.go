package main

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/elazarl/goproxy"
)

var caCert = []byte(`-----BEGIN CERTIFICATE-----
MIIGTzCCBDegAwIBAgIUSXM4670oF/jEBUYRJtZXeEyEWlQwDQYJKoZIhvcNAQEL
BQAwgbYxCzAJBgNVBAYTAkJSMRQwEgYDVQQIDAtNYXJhbmjDg8KjbzEUMBIGA1UE
BwwLU8ODwqNvIEx1aXMxFDASBgNVBAoMC0dydXBvTWF0ZXVzMRQwEgYDVQQLDAtH
cnVwb01hdGV1czEfMB0GA1UEAwwWd3d3LmdydXBvbWF0ZXVzLmNvbS5icjEuMCwG
CSqGSIb3DQEJARYfdGhpYWdvLm1hbGFxdWlhc0BtYXRldXNtYWlzLmNvbTAeFw0y
MjAzMDkxNzQxMDRaFw00MjAzMDQxNzQxMDRaMIG2MQswCQYDVQQGEwJCUjEUMBIG
A1UECAwLTWFyYW5ow4PCo28xFDASBgNVBAcMC1PDg8KjbyBMdWlzMRQwEgYDVQQK
DAtHcnVwb01hdGV1czEUMBIGA1UECwwLR3J1cG9NYXRldXMxHzAdBgNVBAMMFnd3
dy5ncnVwb21hdGV1cy5jb20uYnIxLjAsBgkqhkiG9w0BCQEWH3RoaWFnby5tYWxh
cXVpYXNAbWF0ZXVzbWFpcy5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQDJhxYb9Js1N+DDBCdn54Cb3c9vvy4hhXHDXFbqbeMOACpwtyhj8mbiPlju
vscv6MItoZsbSLbnFd95XT2l0rjrV/iYLTol2Z98KJ3klCP+CFHtBG79u2RlqbSj
0HB0joxdRC9tDdjbPqxUYKY9dOU9tArN1wkvW279y8GqML3ScAH9GeAuTj3ORri+
RoBuBDm/RO9oDPa+q2Uhr9FZsUPsRKfSf/Olrfmo3bDsxPH0Hq4PP075XljPt0vR
WI6p007Z2SuNQcJCDsP5u16UjWBYCyU0QneUdS8UM0udgqMdAnKW/JK52LSJjplz
pTuBIJIGERk+B+CFNYD5vAlgoBt1YJ+6DbiiUIy1HP6FqB4igijtImEHalUmjBbR
JyX/GxCMHDWBNrmr556ntK/oEqw2aokgHUzbpvWFc8yIwRtphxPSEdUkZVsa9Sf6
gGfSQXZjS2KawzpkJb96L2FZZf51vIhiBeR9HFHCXPiMwoyVRzCzsA2IyCVrV9nk
oyB7GNTtzcpMH9J6QOvfbmJasZ0Nhaupz5sgPhxoz6+HuzEp3jFygmSCRi2MTK7T
GXj4W0QrWPT7mkwqe+M22HvGwaC99+J1Xnv8WDg18gy2Dnrh5TOUp86CEhGVAzpE
UR1BTACcudkMJ20MAf+L7JIRE75bftNDtRUvjFNckptoH+9hawIDAQABo1MwUTAd
BgNVHQ4EFgQUzkVhALgmVtuewwXo6cUuX0Is8BowHwYDVR0jBBgwFoAUzkVhALgm
VtuewwXo6cUuX0Is8BowDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOC
AgEAfg3J+KjO56EhWOhK4bfeegqcYokkLAM02xpD5XbSqpxgzjwVlXKXNDWDteA3
GmMXe6voX6unpntFVavx2fsbUz4THGnYaXxar1UDAf0h2tvNXDTzTgIMXusaFV1M
2EoxFNtWP4EetvrsXAEPRAKfxYHc4j4s4i3BlyV9weokWbz/SfqAYCS+oPjATDWY
pWuBHo5fSx8YoCf3WlDSYyxphG5uRjGriump78Oq8TObXXsLOPOhLGUP7zdK/N22
/9RCbIHIDnBvWKh8a0oWZIHpEhkJeO8aafDVexrA1rG7f1bRiXHWPnejw/kPkiwv
1bIu+WZD/7PNBV87VpkcThO+SnLJrlXKmlsI9lDLRt8PQYt5C2Wj4vXAbFFilqFw
z+yG65sVLFcGlkBWTKFMOIAM/PzdYNLA6odJ5zC58MlZHDbibz6PH4/4OOVyLVOU
3CjGfgZNgEGa901ODhstTeAO6kSEIPOw+tVeBFgC+Ugkz2Dp4chpNMyS6Tg+20vN
yKyDeczKrXN+PB2bWSriKzJeu+lquPAHotLXlYLuWIosWa8BfOOGgi9oJFZoE+ig
yz/gUvN9RomccapL8dBWZWyn/u8qVqMoO6O6UWwPeYJ+XVk1KYGeyGee8zNLPBmh
5U0URwVGRLU0RTAoKrcPTGa60OeUFvQusE8vxV1mUgEaDag=
-----END CERTIFICATE-----`)

var caKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`)

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
