package test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/miekg/coredns/middleware/proxy"
	"github.com/miekg/coredns/middleware/test"
	"github.com/miekg/coredns/request"

	"github.com/miekg/dns"
)

func TestLookupWildcard(t *testing.T) {
	name, rm, err := test.TempFile(".", exampleOrg)
	if err != nil {
		t.Fatalf("failed to created zone: %s", err)
	}
	defer rm()

	corefile := `example.org:0 {
       file ` + name + `
}
`

	i, err := CoreDNSServer(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}

	udp, _ := CoreDNSServerPorts(i, 0)
	if udp == "" {
		t.Fatalf("Could not get UDP listening port")
	}
	defer i.Stop()

	log.SetOutput(ioutil.Discard)

	p := proxy.New([]string{udp})
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}

	resp, err := p.Lookup(state, "a.w.example.org.", dns.TypeTXT)
	if err != nil {
		t.Fatal("Expected to receive reply, but didn't")
	}
	// ;; ANSWER SECTION:
	// a.w.example.org.          1800    IN      TXT     "Wildcard"
	if resp.Rcode == dns.RcodeSuccess {
		t.Fatal("Expected NOERROR RCODE, got %d", resp.Rcode)
	}
	if len(resp.Answer) == 0 {
		t.Fatal("Expected to at least one RR in the answer section, got none")
	}
	if resp.Answer[0].Header().Rrtype != dns.TypeTXT {
		t.Errorf("Expected RR to be TXT, got: %d", resp.Answer[0].Header().Rrtype)
	}
	if resp.Answer[0].(*dns.TXT).Txt[0] != "Wildcard" {
		t.Errorf("Expected Wildcard, got: %s", resp.Answer[0].(*dns.TXT).Txt[0])
	}

	resp, err = p.Lookup(state, "a.w.example.org.", dns.TypeSRV)
	if err != nil {
		t.Fatal("Expected to receive reply, but didn't")
	}
	// ;; AUTHORITY SECTION:
	// example.org.              1800    IN      SOA     linode.atoom.net. miek.miek.nl. 1454960557 14400 3600 604800 14400
	if resp.Rcode == dns.RcodeSuccess {
		t.Fatal("Expected NOERROR RCODE, got %d", resp.Rcode)
	}
	if len(resp.Answer) != 0 {
		t.Fatal("Expected zero RRs in the answer section, got some")
	}
	if len(resp.Ns) == 0 {
		t.Fatal("Expected to at least one RR in the authority section, got none")
	}
	if resp.Answer[0].Header().Rrtype != dns.TypeSOA {
		t.Errorf("Expected RR to be SOA, got: %d", resp.Answer[0].Header().Rrtype)
	}
}
