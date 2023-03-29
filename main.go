package main

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
	"log"
	"net"
	"time"
)

var c = cache.New(5*time.Minute, 10*time.Minute)

func handler(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	for _, q := range r.Question {
		n, found := c.Get(q.Name)
		if found {
			c.Set(q.Name, n.(int)+1, cache.DefaultExpiration)
		} else {
			c.Set(q.Name, 1, cache.DefaultExpiration)
			n = 1
		}
		var rr dns.RR
		switch q.Qtype {
		case dns.TypeA:
			var ip net.IP
			{
				if n.(int)%3 == 0 || n.(int) == 1 {
					ip = net.ParseIP("127.0.0.1").To4()
				} else {
					ip = net.ParseIP("66.66.66.66").To4()
				}
				rr = &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					A: ip,
				}
				fmt.Printf("%s ------> %s\n", q.Name, ip)
			}
		case dns.TypeAAAA:
			rr = &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				AAAA: net.ParseIP("::1").To16(),
			}
		default:
			continue
		}
		msg.Answer = append(msg.Answer, rr)
	}
	w.WriteMsg(&msg)
}

func main() {
	dns.HandleFunc(".", handler)
	err := dns.ListenAndServe(":53", "udp", nil)
	if err != nil {
		log.Println(err)
	}
}
