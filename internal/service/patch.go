package service

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/pjcalvo/rigo/internal/config"
)

type PatchService struct {
	interceptConfig config.Config
}

func newPatchService(c config.Config) PatchService {
	return PatchService{
		interceptConfig: c,
	}
}

// intercept handles the logic to match and return the proper response
// should split more
func shouldPatch(uri string, intercepts []config.Intercept) (ok bool, status int, body []byte) {
	for _, intercept := range intercepts {
		if intercept.Match.Uri == "" {
			return
		}

		matched, err := regexp.MatchString(intercept.Match.Uri, uri)
		if err != nil {
			return
		}
		if matched {
			switch intercept.Patch.Type {
			case config.BodyTypeFile:
				body, err = ioutil.ReadFile(intercept.Patch.Body)
				if err != nil {
					return
				}
			case config.BodyTypeString, config.BodyTypeJson:
				body = []byte(intercept.Patch.Body)
				// override the body with the content file
			}

			if intercept.Patch.Status != 0 {
				status = intercept.Patch.Status
			}

			return true, status, body
		}
	}
	return
}

func (i PatchService) HandleRequest(w http.ResponseWriter, r *http.Request) bool {
	if ok, status, body := shouldPatch(r.RequestURI, i.interceptConfig.Intercept.Requests); ok {
		// Handle the intercepted request and return a custom response.
		fmt.Printf("Patching REQUEST for: %s\n	status: %v\n", r.RequestURI, status)
		w.WriteHeader(status)
		w.Write(body)
		return true
	}
	return false
}

func (i PatchService) HandleResponse(r *http.Response) {
	if ok, status, body := shouldPatch(r.Request.URL.String(), i.interceptConfig.Intercept.Responses); ok {
		// Handle the intercepted request and return a custom response.
		// Tood: implement logging
		fmt.Printf("Patching RESPONSE for: %s\n	status: %v\n", r.Request.URL.String(), status)
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.StatusCode = status
	}
}