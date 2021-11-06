package parse_curl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/mattn/go-shellwords"
	"strings"
)

type Header map[string]string

type Request struct {
	Method string `json:"method"`
	Url    string `json:"url"`
	Header Header `json:"header"`
	Body   string `json:"body"`
}

func (r *Request) ToJson(format bool) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	if format {
		encoder.SetIndent("", "  ")
	}
	_ = encoder.Encode(r)

	return string(buffer.Bytes())
}

func Parse(curl string) (*Request, bool) {
	if strings.Index(curl, "curl ") != 0 {
		return nil, false
	}

	// https://github.com/mattn/go-shellwords
	// https://github.com/tj/parse-curl.js
	args, err := shellwords.Parse(curl)
	if err != nil {
		return nil, false
	}

	args = rewrite(args)
	request := &Request{Method: "GET", Header: Header{}}
	state := ""

	for _, arg := range args {
		switch true {
		case isUrl(arg):
			request.Url = arg
			break

		case arg == "-A" || arg == "--user-agent":
			state = "user-agent"
			break

		case arg == "-H" || arg == "--header":
			state = "header"
			break

		case arg == "-d" || arg == "--data" || arg == "--data-ascii" || arg == "--data-raw":
			state = "data"
			break

		case arg == "-u" || arg == "--user":
			state = "user"
			break

		case arg == "-I" || arg == "--head":
			request.Method = "HEAD"
			break

		case arg == "-X" || arg == "--request":
			state = "method"
			break

		case arg == "-b" || arg == "--cookie":
			state = "cookie"
			break

		case len(arg) > 0:
			switch state {
			case "header":
				fields := parseField(arg)
				request.Header[fields[0]] = strings.TrimSpace(fields[1])
				state = ""
				break

			case "user-agent":
				request.Header["User-Agent"] = arg
				state = ""
				break

			case "data":
				if request.Method == "GET" || request.Method == "HEAD" {
					request.Method = "POST"
				}

				if !hasContentType(*request) {
					request.Header["Content-Type"] = "application/x-www-form-urlencoded"
				}

				if len(request.Body) == 0 {
					request.Body = arg
				} else {
					request.Body = request.Body + "&" + arg
				}

				state = ""
				break

			case "user":
				request.Header["Authorization"] = "Basic " +
					base64.StdEncoding.EncodeToString([]byte(arg))
				state = ""
				break

			case "method":
				request.Method = arg
				state = ""
				break

			case "cookie":
				request.Header["Cookie"] = arg
				state = ""
				break

			default:
				break
			}
		}

	}

	// format json body
	if value, ok := request.Header["Content-Type"]; ok && value == "application/json" {
		decoder := json.NewDecoder(strings.NewReader(request.Body))
		jsonData := make(map[string]interface{})
		if err := decoder.Decode(&jsonData); err == nil {
			buffer := &bytes.Buffer{}
			encoder := json.NewEncoder(buffer)
			encoder.SetEscapeHTML(false)
			if err = encoder.Encode(jsonData); err == nil {
				request.Body = strings.ReplaceAll(buffer.String(), "\n", "")
			}
		}
	}

	return request, true
}

func rewrite(args []string) []string {
	res := make([]string, 0)

	for _, arg := range args {

		arg = strings.TrimSpace(arg)

		if arg == "\n" {
			continue
		}

		if strings.Contains(arg, "\n") {
			arg = strings.ReplaceAll(arg, "\n", "")
		}

		// split request method
		if strings.Index(arg, "-X") == 0 {
			res = append(res, arg[0:2])
			res = append(res, arg[2:])
		} else {
			res = append(res, arg)
		}
	}

	return res
}

func isUrl(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://")
}

func parseField(arg string) []string {
	index := strings.Index(arg, ":")
	return []string{arg[0:index], arg[index+2:]}
}

func hasContentType(request Request) bool {
	if _, ok := request.Header["Content-Type"]; ok {
		return true
	}

	return false
}
