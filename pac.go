package sidecar

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/pmezard/adblock/adblock"
)

type Pac struct {
	Matcher *adblock.RuleMatcher
}

func NewPac(cfg *Config) *Pac {
	p := &Pac{
		Matcher: nil,
	}
	p.getGfwList(cfg.Server, cfg.ComplexPath, cfg.CustomHeaders, cfg.GfwListUrl)
	return p
}

func (p *Pac) getGfwList(server string, subpath string, headers map[string]string, url string) {
	//url: https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic("fetch gfwlist failed.")
	}
	req_url := req.URL
	raw_path := req_url.Path
	req_url.Host = server
	req_url.Path = "/" + subpath + "/" + req.Host + raw_path
	req.Host = server
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		panic("fetch gfwlist failed.")
	}
	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	matcher := adblock.NewMatcher()
	rules, _ := adblock.ParseRules(decoder)
	for _, rule := range rules {
		matcher.AddRule(rule, 0)
	}
	p.Matcher = matcher
}

func (p *Pac) Compare(req *http.Request) bool {
	url := req.URL
	url.Scheme = "https"
	matched, _, err := p.Matcher.Match(
		&adblock.Request{
			URL:     url.String(),
			Timeout: 10 * time.Millisecond,
		})
	if err != nil {
		return false
	}
	return matched
}
