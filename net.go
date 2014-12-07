package vkapi

import (
  "bytes"
  "fmt"
  "io"
  "log"
  "mime/multipart"
  "net/http"
  "os"
  "path/filepath"

  "github.com/demdxx/gocast"
)

func newUploadRequest(uri string, params map[string]interface{}, paramName, path string, rbody io.Reader) (*http.Request, error) {
  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)

  if nil != rbody {
    part, err := writer.CreateFormFile(paramName, filepath.Base(path))
    if err != nil {
      return nil, err
    }
    _, err = io.Copy(part, rbody)
  }

  for key, val := range params {
    _ = writer.WriteField(key, gocast.ToString(val))
  }
  err = writer.Close()
  if err != nil {
    return nil, err
  }

  return http.NewRequest("POST", uri, body)
}

func newPostRequest(uri string, params map[string]interface{}) (*http.Request, error) {
  return newUploadRequest(uri, params, "", "", nil)
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]interface{}, paramName, path string) (*http.Request, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  return newUploadRequest(uri, params, paramName, path, file)
}

func build_query(params map[string]interface{}) string {
  query := make([]string, 0)
  if nil != params {
    for k, v := range params {
      query = append(query, k+"="+url.QueryEscape(gocast.ToString(v)))
    }
  }
  return strings.Join(query, "&")
}
