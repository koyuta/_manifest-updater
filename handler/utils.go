package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// RetrieveBody unmarshals req object from retrieved http request body.
func RetrieveBody(body io.ReadCloser, req interface{}) error {
	b, err := ioutil.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, &req)
}
