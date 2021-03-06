/* 
 * A name server which sends back the IP address of its client, the
 * recursive resolver. When queried for type TXT, it sends back the text
 * form of the address.  When queried for type A (resp. AAAA), it sends
 * back the IPv4 (resp. v6) address.
 *
 * Similar services: whoami.ultradns.net, whoami.akamai.net. Also (but it
 * is not their normal goal): rs.dns-oarc.net, porttest.dns-oarc.net,
 * amiopen.openresolvers.org.
 *
 * Stephane Bortzmeyer <stephane+grong@bortzmeyer.org>
 *
 * Miek Gieben <miek@miek.nl>
 * Adapted to Go DNS (i.e. completely rewritten)
 */

package main

import (
	"net"
	"dns"
	"fmt"
	"os/signal"
	"strconv"
)

var dom = "whoami.miek.nl"

func handleReflect(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Extra = make([]dns.RR, 1)
	m.Answer = make([]dns.RR, 1)
        var (
                v4 bool
                rr dns.RR
                str string
                a net.IP
        )

        if ip, ok := w.RemoteAddr().(*net.UDPAddr); ok {
                str = "Port: " + strconv.Itoa(ip.Port) + " (udp)"
                a = ip.IP
                v4 = a.To4() != nil
        }
        if ip, ok := w.RemoteAddr().(*net.TCPAddr); ok {
                str = "Port: " + strconv.Itoa(ip.Port) + " (tcp)"
                a = ip.IP
                v4 = a.To4() != nil
        }

	if v4 {
		rr = new(dns.RR_A)
                rr.(*dns.RR_A).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.RR_A).A = a
	} else {
		rr = new(dns.RR_AAAA)
                rr.(*dns.RR_AAAA).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.RR_AAAA).AAAA = a
	}

	t := new(dns.RR_TXT)
	t.Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	t.Txt = str
	m.Extra[0] = t
	m.Answer[0] = rr

	b, ok := m.Pack()
        if !ok { return }
	w.Write(b)
}

func serve(net string) {
        err := dns.ListenAndServe(":8053", net, nil)
        if err != nil {
                fmt.Printf("Failed to setup the " + net + " server: %s\n", err.String())
        }
}

func main() {
	dns.HandleFunc(".", handleReflect)
        go serve("udp4")
        go serve("udp6")
        go serve("tcp4")
        go serve("tcp6")
forever:
        for {
                select {
                case <-signal.Incoming:
                        fmt.Printf("Signal received, stopping\n")
                        break forever
                }
        }
}
