// The MIT License (MIT)
//
// Copyright (c) 2014 Dmitry Ponomarev <demdxx@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package vkapi

import (
  "crypto/md5"
  "errors"
  "fmt"
  "io"

  "io/ioutil"

  "net/http"

  "encoding/json"
  "encoding/xml"
)

const OAUTH_URL string = "https://oauth.vk.com"
const AUTHORIZE_URL string = "/authorize?"
const ACCESS_TOKEN_URL string = "/access_token?"
const TOKEN_URL string = "/token?"
const API_URL string = "https://api.vk.com"

type Vk struct {
  Client      *http.Client
  AccessToken string
  ClientId    string
  Secret      string
  Params      map[string]interface{}
  Format      string
}

func MakeVk(client *http.Client, access_token, client_id, secret string, params map[string]interface{}) *Vk {
  return &Vk{
    Client:      client,
    AccessToken: access_token,
    ClientId:    client_id,
    Secret:      secret,
    Params:      params,
    Format:      "",
  }
}

///////////////////////////////////////////////////////////////////////////////
/// Params
///////////////////////////////////////////////////////////////////////////////

func (vk *Vk) SetJsonFormat() *Vk {
  vk.Format = ""
  return vk
}

func (vk *Vk) SetXmlFormat() *Vk {
  vk.Format = ".xml"
  return vk
}

func (vk *Vk) IsJsonResponse() bool {
  return 0 == len(vk.Format)
}

func (vk *Vk) ApiVersion() string {
  if nil != vk.Params {
    if v, ok := vk.Params["v"]; ok && 0 != len(v.(string)) {
      return v.(string)
    }
  }
  return "5.1"
}

///////////////////////////////////////////////////////////////////////////////
/// Actions
///////////////////////////////////////////////////////////////////////////////

func (vk *Vk) AuthToken() error {
  params := map[string]interface{}{
    "grant_type":    "client_credentials",
    "client_id":     vk.ClientId,
    "client_secret": vk.Secret,
    "v":             vk.ApiVersion(),
  }

  auth_url := OAUTH_URL + ACCESS_TOKEN_URL + build_query(params)

  var response map[string]interface{}
  err := vk.get_url(auth_url, &response)

  if nil != err {
    if e, ok := response["error"]; ok {
      err = errors.New(e.(map[string]interface{})["'error_description'"].(string))
    } else {
      vk.AccessToken = response["access_token"].(string)
    }
  }
  return err
}

// Auth by login/password
func (vk *Vk) AuthDirect(username, password, scope, test_redirect_uri string) (map[string]interface{}, error) {
  params := map[string]interface{}{
    "grant_type":        "password",
    "client_id":         vk.ClientId,
    "client_secret":     vk.Secret,
    "username":          username,
    "password":          password,
    "scope":             scope,
    "test_redirect_uri": test_redirect_uri,
    "v":                 vk.ApiVersion(),
  }

  // Prepare request url
  request_url := OAUTH_URL + TOKEN_URL + build_query(params)

  var response map[string]interface{}
  err := vk.get_url(request_url, &response)

  return response, err
}

// Request api GET method
func (vk *Vk) Api(method string, params map[string]interface{}, response interface{}) error {
  prms := vk.prepare_params(params)

  // Prepare request url
  request_url := "/method/" + method + vk.Format + "?" + build_query(prms)

  if len(vk.Secret) > 0 {
    sig := md5_s(request_url + vk.Secret)
    request_url = API_URL + request_url + "&sig=" + sig
  }
  return vk.get_url(request_url, response)
}

// Request api POST method
func (vk *Vk) ApiPost(method string, params map[string]interface{}, postParams map[string]interface{}, response interface{}) error {
  prms := vk.prepare_params(params)

  // Prepare request url
  request_url := "/method/" + method + vk.Format + "?" + build_query(prms)

  if len(vk.Secret) > 0 {
    sig := md5_s(request_url + vk.Secret)
    request_url = API_URL + request_url + "&sig=" + sig
  }
  return vk.post_url(request_url, nil, response)
}

// Request api POST file method
func (vk *Vk) ApiPostFile(method, filepath string, params map[string]interface{}, postParams map[string]interface{}, response interface{}) error {
  prms := vk.prepare_params(params)

  // Prepare request url
  request_url := "/method/" + method + vk.Format + "?" + build_query(prms)

  if len(vk.Secret) > 0 {
    sig := md5_s(request_url + vk.Secret)
    request_url = API_URL + request_url + "&sig=" + sig
  }
  return vk.postFile_url(request_url, filepath, postParams, response)
}

// Request api POST file method by byte body
func (vk *Vk) ApiPostFileBody(method, filename string, body io.Reader, params map[string]interface{}, postParams map[string]interface{}, response interface{}) error {
  prms := vk.prepare_params(params)

  // Prepare request url
  request_url := "/method/" + method + vk.Format + "?" + build_query(prms)

  if len(vk.Secret) > 0 {
    sig := md5_s(request_url + vk.Secret)
    request_url = API_URL + request_url + "&sig=" + sig
  }
  return vk.postFileBody_url(request_url, filename, postParams, body, response)
}

func (vk *Vk) RawGet(url string, response interface{}) error {
  return vk.get_url(url, response)
}

func (vk *Vk) RawPostFile(url, filepath string, postParams map[string]interface{}, response interface{}) error {
  return vk.postFile_url(url, filepath, postParams, response)
}

func (vk *Vk) RawPostFileBody(url, filepath string, postParams map[string]interface{}, body io.Reader, response interface{}) error {
  return vk.postFile_url(url, filepath, postParams, response)
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func (vk *Vk) get_url(url string, response interface{}) error {
  // Send request
  resp, err := vk.Client.Get(url)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  data, err := ioutil.ReadAll(resp.Body)

  // Process response
  if nil == err {
    if vk.IsJsonResponse() {
      json.Unmarshal(data, response)
    } else {
      xml.Unmarshal(data, response)
    }
    response = prepareResponse(response)
  }
  return err
}

func (vk *Vk) post_url(url string, params map[string]interface{}, response interface{}) error {
  req, err := newPostRequest(url, params)
  if nil != err {
    return err
  }
  return vk.processRequest(req, response)
}

func (vk *Vk) postFile_url(url, filepath string, params map[string]interface{}, response interface{}) error {
  req, err := newfileUploadRequest(url, params, "file1", filepath)
  if nil != err {
    return err
  }
  return vk.processRequest(req, response)
}

func (vk *Vk) postFileBody_url(url, filepath string, params map[string]interface{}, body io.Reader, response interface{}) error {
  req, err := newUploadRequest(url, params, "file1", filepath, body)
  if nil != err {
    return err
  }
  return vk.processRequest(req, response)
}

func (vk *Vk) processRequest(req *http.Request, response interface{}) error {
  // Send request
  resp, err := vk.Client.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  data, err := ioutil.ReadAll(resp.Body)

  // Process response
  if nil == err {
    if vk.IsJsonResponse() {
      json.Unmarshal(data, response)
    } else {
      xml.Unmarshal(data, response)
    }
    *response.(*interface{}) = prepareResponse(response)
  }
  return err
}

func (vk *Vk) prepare_params(params map[string]interface{}) map[string]interface{} {
  var prms = make(map[string]interface{})
  if len(vk.AccessToken) > 0 {
    prms["access_token"] = vk.AccessToken
  }

  // prepare params
  if nil != params {
    for k, v := range params {
      prms[k] = v
    }
  }

  if nil != vk.Params {
    for k, v := range vk.Params {
      prms[k] = v
    }
  }
  return prms
}

///////////////////////////////////////////////////////////////////////////////
/// Nonclass Helpers
///////////////////////////////////////////////////////////////////////////////

func prepareResponse(resp interface{}) map[string]interface{} {
  fmt.Println("prepareResponse", resp)
  if nil != resp {
    switch resp.(type) {
    case map[string]interface{}:
      return resp.(map[string]interface{})
    case map[interface{}]interface{}:
      data := make(map[string]interface{})
      for k, v := range resp.(map[interface{}]interface{}) {
        data[k.(string)] = v
      }
    case map[string]string:
      data := make(map[string]interface{})
      for k, v := range resp.(map[string]string) {
        data[k] = v
      }
    }
  }
  return nil
}

func md5_s(text string) string {
  h := md5.New()
  io.WriteString(h, text)
  return fmt.Sprintf("%x", h.Sum(nil))
}
