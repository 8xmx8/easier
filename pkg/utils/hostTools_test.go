package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchRootDomain(t *testing.T) {
	type caseT struct {
		name   string
		url    string
		domain string
	}

	cases := []*caseT{
		{
			name:   "domain",
			url:    "http://google.com",
			domain: "google.com",
		},
		{
			name:   "domain",
			url:    "http://blog.google",
			domain: "blog.google",
		},
		{
			name:   "domain",
			url:    "www.medi-cal.ca.gov/",
			domain: "ca.gov",
		},
		{
			name:   "domain",
			url:    "https://ato.gov.au",
			domain: "ato.gov.au",
		},
		{
			name:   "domain",
			url:    "http://a.very.complex-domain.co.uk:8080/foo/bar",
			domain: "complex-domain.co.uk",
		},
		{
			name:   "domain",
			url:    "http://a.domain.that.is.unmanaged",
			domain: "is.unmanaged",
		},
		{
			name:   "domain",
			url:    "bhhthl.com/index.php?s=/Extend/guestbook.html",
			domain: "bhhthl.com",
		},
		{
			name:   "ip",
			url:    "103.54.45.103/1/",
			domain: "103.54.45.103",
		},
		{
			name:   "ip",
			url:    "http://143.250.236.3/",
			domain: "143.250.236.3",
		},
	}

	for _, cc := range cases {
		actual, err := MatchRootDomain(cc.url)
		assert.NoError(t, err)
		assert.Equal(t, cc.domain, actual)
	}
}
