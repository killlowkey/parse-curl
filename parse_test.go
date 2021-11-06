package parse_curl

import (
	"github.com/bmizerany/assert"
	"testing"
)

type M map[string]interface{}

func TestParse(t *testing.T) {

	addSample(t, "curl -XPUT http://api.sloths.com/sloth/4", M{
		"method": "PUT",
		"url":    "http://api.sloths.com/sloth/4",
	})

	addSample(t, "curl http://api.sloths.com", M{
		"method": "GET",
		"url":    "http://api.sloths.com",
	})

	addSample(t, "curl -H \"Accept-Encoding: gzip\" --compressed http://api.sloths.com", M{
		"method": "GET",
		"url":    "http://api.sloths.com",
		"header": M{
			"Accept-Encoding": "gzip",
		},
	})

	addSample(t, "curl -X DELETE http://api.sloths.com/sloth/4", M{
		"method": "DELETE",
		"url":    "http://api.sloths.com/sloth/4",
	})

	addSample(t, "curl -d \"foo=bar\" https://api.sloths.com", M{
		"method": "POST",
		"url":    "https://api.sloths.com",
		"header": M{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		"body": "foo=bar",
	})

	addSample(t, "curl -H \"Accept: text/plain\" --header \"User-Agent: slothy\" https://api.sloths.com", M{
		"method": "GET",
		"url":    "https://api.sloths.com",
		"header": M{
			"Accept":     "text/plain",
			"User-Agent": "slothy",
		},
	})

	addSample(t, "curl --cookie 'species=sloth;type=galactic' slothy https://api.sloths.com", M{
		"method": "GET",
		"url":    "https://api.sloths.com",
		"header": M{
			"Cookie": "species=sloth;type=galactic",
		},
	})

	addSample(t, "curl --location --request GET 'http://api.sloths.com/users?token=admin'", M{
		"method": "GET",
		"url":    "http://api.sloths.com/users?token=admin",
	})
}

func addSample(t *testing.T, url string, exp M) {
	request, _ := Parse(url)
	check(t, exp, request)
}

func check(t *testing.T, exp M, got *Request) {
	for key, value := range exp {
		switch key {
		case "method":
			assert.Equal(t, value, got.Method)
		case "url":
			assert.Equal(t, value, got.Url)
		case "body":
			assert.Equal(t, value, got.Body)
		case "header":
			headers := value.(M)
			for k, v := range headers {
				assert.Equal(t, v, got.Header[k])
			}
		}
	}
}
