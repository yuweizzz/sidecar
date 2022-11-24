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
	if cfg.Sidecar.GfwListUrl != "" {
		p.getGfwList(cfg.RemoteProxy.Server, cfg.RemoteProxy.ComplexPath, cfg.RemoteProxy.CustomHeaders, cfg.Sidecar.GfwListUrl)
	}
	if cfg.Sidecar.CustomProxyHosts != nil {
		p.ExpandHosts(cfg.Sidecar.CustomProxyHosts)
	}
	return p
}

func (p *Pac) getGfwList(server string, subpath string, headers map[string]string, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic("Fetch GfwList failed.")
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
		panic("Fetch GfwList failed.")
	}
	Info("Fetch GfwList from ", url)
	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	if p.Matcher == nil {
		p.Matcher = adblock.NewMatcher()
	}
	rules, _ := adblock.ParseRules(decoder)
	for _, rule := range rules {
		p.Matcher.AddRule(rule, 0)
	}
}

func (p *Pac) ExpandHosts(list []string) {
	if p.Matcher == nil {
		p.Matcher = adblock.NewMatcher()
	}
	var rule *adblock.Rule
	var err error
	for _, host := range list {
		rule, err = adblock.ParseRule("||" + host)
		if err != nil {
			Debug("Parse Pac Rule failed, host: ", host)
			continue
		}
		p.Matcher.AddRule(rule, 0)
		Debug("Parse Pac Rule host: ", host)
	}
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
