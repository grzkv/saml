// This is an example that implements a bitly-esque short link service.
package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/metadata"
	"github.com/dchest/uniuri"
	"github.com/kr/pretty"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var links = map[string]Link{}
var samlsp *saml.ServiceProvider

type Link struct {
	ShortLink string
	Target    string
	Owner     string
}

// CreateLink handles requests to create links
func CreateLink(c web.C, w http.ResponseWriter, r *http.Request) {
	account := r.Header.Get("X-Remote-User")
	l := Link{
		ShortLink: uniuri.New(),
		Target:    r.FormValue("t"),
		Owner:     account,
	}
	links[l.ShortLink] = l

	fmt.Fprintf(w, "%s\n", l.ShortLink)
	return
}

// ServeLink handles requests to redirect to a link
func ServeLink(c web.C, w http.ResponseWriter, r *http.Request) {
	l, ok := links[strings.TrimPrefix(r.URL.Path, "/")]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, l.Target, http.StatusFound)
	return
}

// ListLinks returns a list of the current user's links
func ListLinks(c web.C, w http.ResponseWriter, r *http.Request) {
	account := r.Header.Get("X-Remote-User")
	for _, l := range links {
		if l.Owner == account {
			fmt.Fprintf(w, "%s\n", l.ShortLink)
		}
	}
}

func main() {
	baseURL := flag.String("url", "https://962766ce.ngrok.io", "The base URL of this service")
	idpMetadataURL := flag.String("idp", "https://516becc2.ngrok.io/metadata", "The metadata URL for the IDP")
	flag.Parse()

	samlsp = &saml.ServiceProvider{
		MetadataURL: *baseURL + "/saml/metadata",
		AcsURL:      *baseURL + "/saml/acs",
	}
	samlsp.Key = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDU8wdiaFmPfTyRYuFlVPi866WrH/2JubkHzp89bBQopDaLXYxi
3PTu3O6Q/KaKxMOFBqrInwqpv/omOGZ4ycQ51O9I+Yc7ybVlW94lTo2gpGf+Y/8E
PsVbnZaFutRctJ4dVIp9aQ2TpLiGT0xX1OzBO/JEgq9GzDRf+B+eqSuglwIDAQAB
AoGBAMuy1eN6cgFiCOgBsB3gVDdTKpww87Qk5ivjqEt28SmXO13A1KNVPS6oQ8SJ
CT5Azc6X/BIAoJCURVL+LHdqebogKljhH/3yIel1kH19vr4E2kTM/tYH+qj8afUS
JEmArUzsmmK8ccuNqBcllqdwCZjxL4CHDUmyRudFcHVX9oyhAkEA/OV1OkjM3CLU
N3sqELdMmHq5QZCUihBmk3/N5OvGdqAFGBlEeewlepEVxkh7JnaNXAXrKHRVu/f/
fbCQxH+qrwJBANeQERF97b9Sibp9xgolb749UWNlAdqmEpmlvmS202TdcaaT1msU
4rRLiQN3X9O9mq4LZMSVethrQAdX1whawpkCQQDk1yGf7xZpMJ8F4U5sN+F4rLyM
Rq8Sy8p2OBTwzCUXXK+fYeXjybsUUMr6VMYTRP2fQr/LKJIX+E5ZxvcIyFmDAkEA
yfjNVUNVaIbQTzEbRlRvT6MqR+PTCefC072NF9aJWR93JimspGZMR7viY6IM4lrr
vBkm0F5yXKaYtoiiDMzlOQJADqmEwXl0D72ZG/2KDg8b4QZEmC9i5gidpQwJXUc6
hU+IVQoLxRq0fBib/36K9tcrrO5Ba4iEvDcNY+D8yGbUtA==
-----END RSA PRIVATE KEY-----
`
	samlsp.Certificate = `-----BEGIN CERTIFICATE-----
MIIB7zCCAVgCCQDFzbKIp7b3MTANBgkqhkiG9w0BAQUFADA8MQswCQYDVQQGEwJV
UzELMAkGA1UECAwCR0ExDDAKBgNVBAoMA2ZvbzESMBAGA1UEAwwJbG9jYWxob3N0
MB4XDTEzMTAwMjAwMDg1MVoXDTE0MTAwMjAwMDg1MVowPDELMAkGA1UEBhMCVVMx
CzAJBgNVBAgMAkdBMQwwCgYDVQQKDANmb28xEjAQBgNVBAMMCWxvY2FsaG9zdDCB
nzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEA1PMHYmhZj308kWLhZVT4vOulqx/9
ibm5B86fPWwUKKQ2i12MYtz07tzukPymisTDhQaqyJ8Kqb/6JjhmeMnEOdTvSPmH
O8m1ZVveJU6NoKRn/mP/BD7FW52WhbrUXLSeHVSKfWkNk6S4hk9MV9TswTvyRIKv
Rsw0X/gfnqkroJcCAwEAATANBgkqhkiG9w0BAQUFAAOBgQCMMlIO+GNcGekevKgk
akpMdAqJfs24maGb90DvTLbRZRD7Xvn1MnVBBS9hzlXiFLYOInXACMW5gcoRFfeT
QLSouMM8o57h0uKjfTmuoWHLQLi6hnF+cvCsEFiJZ4AbF+DgmO6TarJ8O05t8zvn
OwJlNCASPZRH/JmF8tX0hoHuAQ==
-----END CERTIFICATE-----
`

	fmt.Printf("fetching IDP metadata: %s\n", *idpMetadataURL)
	resp, err := http.Get(*idpMetadataURL)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	r := metadata.EntitiesDescriptor{}
	if err := xml.Unmarshal(buf, &r); err != nil {
		panic(err)
	}
	for _, e := range r.EntityDescriptor {
		if e.IDPSSODescriptor != nil {
			samlsp.IDPMetadata = e
			break
		}
	}
	if samlsp.IDPMetadata == nil {
		panic("cannot find idp in metadata")
	}

	// register with the service provider
	spMetadataBuf, _ := xml.MarshalIndent(metadata.EntitiesDescriptor{
		EntityDescriptor: []*metadata.Metadata{samlsp.Metadata()},
	}, "", "  ")
	http.Post(strings.Replace(*idpMetadataURL, "/metadata", "/register-sp", 1), "text/xml", bytes.NewReader(spMetadataBuf))

	samlMiddleware := &saml.ServiceProviderMiddleware{ServiceProvider: samlsp}
	goji.Handle("/saml/*", samlMiddleware)

	authMux := web.New()
	authMux.Use(samlMiddleware.RequireAccountMiddleware)
	authMux.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
		pretty.Fprintf(w, "%# v", r)
	})
	authMux.Post("/", CreateLink)
	authMux.Get("/", ListLinks)

	goji.Handle("/*", authMux)
	goji.Get("/:link", ServeLink)

	goji.Serve()
}
