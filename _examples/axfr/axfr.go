package main

import (
	"fmt"
	"dns"
	"flag"
)

func main() {
	var serial *int = flag.Int("serial", 0, "Perform an IXFR with the given serial")
	var nameserver *string = flag.String("ns", "127.0.0.1:53", "Query this nameserver")
	flag.Parse()
	zone := flag.Arg(flag.NArg() - 1)

        client := dns.NewClient()
        client.Net = "tcp"
        m := new(dns.Msg)
	if *serial > 0 {
		m.SetIxfr(zone, uint32(*serial))
	} else {
		m.SetAxfr(zone)
	}
        axfr, err := client.XfrReceive(m, *nameserver)
        if err != nil {
                println(err.String())
        }
        for _, v := range axfr {
		fmt.Printf("%v\n", v)
	}
}
