package dns

import (
	"os"
)

// Perform an incoming Ixfr or Axfr. If the message q's question
// section contains an AXFR type an Axfr is performed. If q's question
// section contains an IXFR type an Ixfr is performed.
func (c *Client) XfrReceive(q *Msg, a string) ([]*Msg, os.Error) {
	w := new(reply)
	w.client = c
	w.addr = a
	w.req = q // is this needed TODO(mg)

	if err := w.Send(q); err != nil {
		return nil, err
	}
	// conn should be set now
	switch q.Question[0].Qtype {
	case TypeAXFR:
		return w.axfrReceive()
	case TypeIXFR:
		//	return w.ixfrReceive()
	}
	panic("not reached")
	return nil, nil
}

func (w *reply) axfrReceive() ([]*Msg, os.Error) {
	axfr := make([]*Msg, 0) // use append ALL the time?
	first := true
	for {
		in, err := w.Receive()
		axfr = append(axfr, in)
		if err != nil {
			return axfr, err
		}

		if first {
			if !checkXfrSOA(in, true) {
				return axfr, ErrXfrSoa
			}
			first = !first
		}

		if !first {
			w.tsigTimersOnly = true // Subsequent envelopes use this.
			if !checkXfrSOA(in, false) {
				// Soa record not the last one
				continue
			} else {
				return axfr, nil
			}
		}
	}
	panic("not reached")
	return nil, nil
}
/*

// Perform an outgoing Ixfr or Axfr. If the message q's question
// section contains an AXFR type an Axfr is performed. If q's question
// section contains an IXFR type an Ixfr is performed.
// The actual records to send are given on the channel m. And errors
// during transport are return on channel e.
func (d *Conn) XfrWrite(q *Msg, m chan *Xfr, e chan os.Error) {
	switch q.Question[0].Qtype {
	case TypeAXFR:
		d.axfrWrite(q, m, e)
	case TypeIXFR:
		//                d.ixfrWrite(q, m)
        default:
                e <- &Error{Error: "Xfr Qtype not recognized"}
                close(m)
	}
}

// Just send the zone
func (d *Conn) axfrWrite(q *Msg, m chan *Xfr, e chan os.Error) {
	out := new(Msg)
	out.Id = q.Id
	out.Question = q.Question
	out.Answer = make([]RR, 1001) // TODO(mg) look at this number
	out.MsgHdr.Response = true
	out.MsgHdr.Authoritative = true
        first := true
	var soa *RR_SOA
	i := 0
	for r := range m {
		out.Answer[i] = r.RR
		if soa == nil {
			if r.RR.Header().Rrtype != TypeSOA {
				e <- ErrXfrSoa
                                return
			} else {
				soa = r.RR.(*RR_SOA)
			}
		}
		i++
		if i > 1000 {
			// Send it
			err := d.WriteMsg(out)
			if err != nil {
				e <- err
                                return
			}
			i = 0
			// Gaat dit goed?
			out.Answer = out.Answer[:0]
                        if first {
                                if d.Tsig != nil {
                                        d.Tsig.TimersOnly = true
                                }
                                first = !first
                        }
		}
	}
	// Everything is sent, only the closing soa is left.
	out.Answer[i] = soa
	out.Answer = out.Answer[:i+1]
	err := d.WriteMsg(out)
	if err != nil {
		e <- err
	}
}

func (d *Conn) ixfrReceive(q *Msg, m chan *Xfr) {
	defer close(m)
	var serial uint32 // The first serial seen is the current server serial
	var x *Xfr
	first := true
	in := new(Msg)
	for {

		err := d.ReadMsg(in)
		if err != nil {
			m <- &Xfr{true, nil, err}
			return
		}
		if in.Id != q.Id {
			m <- &Xfr{true, nil, ErrId}
			return
		}

		if first {
			// A single SOA RR signals "no changes"
			if len(in.Answer) == 1 && checkXfrSOA(in, true) {
				return
			}

			// But still check if the returned answer is ok
			if !checkXfrSOA(in, true) {
				m <- &Xfr{true, nil, ErrXfrSoa}
				return
			}
			// This serial is important
			serial = in.Answer[0].(*RR_SOA).Serial
			first = !first
		}

		// Now we need to check each message for SOA records, to see what we need to do
		x.Add = true
		if !first {
			if d.Tsig != nil {
				d.Tsig.TimersOnly = true
			}
			for k, r := range in.Answer {
				// If the last record in the IXFR contains the servers' SOA,  we should quit
				if r.Header().Rrtype == TypeSOA {
					switch {
					case r.(*RR_SOA).Serial == serial:
						if k == len(in.Answer)-1 {
							// last rr is SOA with correct serial
							//m <- r dont' send it
							return
						}
						x.Add = true
						if k != 0 {
							// Intermediate SOA
							continue
						}
					case r.(*RR_SOA).Serial != serial:
						x.Add = false
						continue // Don't need to see this SOA
					}
				}
				x.RR = r
				m <- x
			}
		}
	}
	panic("not reached")
	return
}
*/

// Check if he SOA record exists in the Answer section of 
// the packet. If first is true the first RR must be a soa
// if false, the last one should be a SOA
func checkXfrSOA(in *Msg, first bool) bool {
	if len(in.Answer) > 0 {
		if first {
			return in.Answer[0].Header().Rrtype == TypeSOA
		} else {
			return in.Answer[len(in.Answer)-1].Header().Rrtype == TypeSOA
		}
	}
	return false
}
