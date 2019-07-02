/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package routes

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

var (
	testdata = map[string]interface{}{
		"id_rsa": []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAv2TkunvS/lUZm7oH4cpHvy/QyT72kwoVdEzBwfBNZoXWmC2h
P/+qoTSXiq6FGwvVLSaiOSIR1WDDuihNXmR4zXDAeObHicVlldmmG8NzgCW5ZjO7
1fnmGXgv76VtrB6ccu5K2JxWSzvLrrH9/PmibukyRyO4hBhxeBhipeaLi2HM2SJS
eu867NzBKbQqbsaXbb+D+Ko4T4C5ouJATYe6+0ZyhqyFJmz6ARoDnNCE5DvPXniA
K0b3Mm/AJbEBX0p8OA6m5xeeRFggueaA3BBw1NOcZoVLDo2XNH5vWrj1eyMZHac7
tirTDOE2VxR14xGhjuYSaGO/yc0VjjtUX8N7VQIDAQABAoIBAHUO7UINR5/cRpxT
LEzxne4V/ZmIU+DcswB9jafjJEPHKdfLWKs+4IpWEzVzxd8j3o8N6PwOlV+vHia2
TZOk2am1A1Muuu3NeHMtOgYTBYpkCD+09nZJsGz1cEQfJrO1yTQWAFr5S2IaQVoo
bNKTj8BMCj8uXsUT+hpct8EF/2UQK37VhfkDOvrRtTVDiw88DJNlcVf9Ugo0GG2h
TZT1RT/h6JeEA1+AomHsSfEJ4XKlmRaHqstHl3T9JPoeNlUr4KgFeVg+cH0ta3EM
8lQsv1z5jVI/v9vBuUN2tgq4OQ/fGl8KWlKZa8XyIhz3sKusfquCFx1vWah3Jm+h
w/UhCSUCgYEA/qbQkISXX3l6QcFq/ojjuH4+YmeGaRmu/F2+oLNB35Xr1AqBrPLH
vT8WE2b+bAahw+my353x1iQygABO3U1KQP+RzKi6CMwrDSNn6y/xyGxIPKR/MdiE
m3Yj+LxsWrTIYfhzPchP5L2qB3V7Dt8WUHhGI3kIsOpm0i5dx9NgVlcCgYEAwGhU
+/T3/6NBY3wUd0ngXlKiDNHt34ZoCoAY0NLY/lsiWUtwYlgC9Qs91cX9aLCyO6mZ
LF8AJ7aaVMyRa0r5ycem9uIzJUhDZo+9v0Zb9fbkjoQ86/BGezJA+79Sy3Q54s+D
jJ9ElaCJ4n6/FIP/IkK+9Mgp5G8Ts8ufEE7C+DMCgYEAt3iUuCrvrRgm341teypB
d8FtTRTtoHeiva0FFV9RzLeFi+Zt+5+IDW+QhjYkhMxabH7KI5b2kKTPxa1zJLr1
DtOTxnKiZohDVFn4G1kVyKNLgHW8Nrua/y8lR6bqIogx/3Q0A4V5GoMUJ/+aw+Iz
f5LIZfJkMqMPpctGQhynQk0CgYEAjE7Iwl92Rc4YTeLazc5qtn3NvEmODHVoA1g8
QHOxV3K/zpwLnTuPFICZG/3geGp53rYjg87XPx6S1onC9ZncI3/bSqfTIjnbJLxn
Y0d8ohXjv+XAw1EZJJeV+b8gMktUNwiaZn5yNia2xhslXmGPOL2xoLEik3lIxdET
8oFs/i8CgYEA8PguaOdYUDyumAwtiq59D8Zs0ZpmWOO59eyvIqalZlNdG04Mmpg1
rNb9zpL+Jy5lj+4NAWjgEaxJWQuWF+Gc8epZk8vw0GNf3VCoZiHSEM4F2u8V7IoN
gTqcfiuPwIa74zsLJnyCsuERYXmkz/almnTrJx0xkPj55CzRfyZX31w=
-----END RSA PRIVATE KEY-----
`),
		"id_rsa.pub": []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv2TkunvS/lUZm7oH4cpH
vy/QyT72kwoVdEzBwfBNZoXWmC2hP/+qoTSXiq6FGwvVLSaiOSIR1WDDuihNXmR4
zXDAeObHicVlldmmG8NzgCW5ZjO71fnmGXgv76VtrB6ccu5K2JxWSzvLrrH9/Pmi
bukyRyO4hBhxeBhipeaLi2HM2SJSeu867NzBKbQqbsaXbb+D+Ko4T4C5ouJATYe6
+0ZyhqyFJmz6ARoDnNCE5DvPXniAK0b3Mm/AJbEBX0p8OA6m5xeeRFggueaA3BBw
1NOcZoVLDo2XNH5vWrj1eyMZHac7tirTDOE2VxR14xGhjuYSaGO/yc0VjjtUX8N7
VQIDAQAB
-----END PUBLIC KEY-----
`),
	}
)

func init() {
	viper.Set("key.private", testdata["id_rsa"])
	viper.Set("key.public", testdata["id_rsa.pub"])
	testdata["key.private"], _ = jwt.ParseRSAPrivateKeyFromPEM(testdata["id_rsa"].([]byte))
	testdata["key.public"], _ = jwt.ParseRSAPublicKeyFromPEM(testdata["id_rsa.pub"].([]byte))

}
