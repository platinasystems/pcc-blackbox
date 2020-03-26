package pcc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

type HttpResp struct {
	Path      string
	Status    int
	Message   string
	Error     string
	RequestId string
	Data      []byte
}

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}
}

type pccRespGeneric struct {
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Error     string `json:"error"`
	RequestId string `json:"requestId"`
	Data      interface{}
}

// The PCC gateway
type RestClient struct {
	address string
	port    int
	token   string
}

// Evaluate and unmarshal the response
func unmarshalHttpResp(url string, resp *HttpResp, out interface{}) (err error) {
	if resp.Status == 200 {
		if out != nil {
			if err = json.Unmarshal(resp.Data, out); err != nil {
				fmt.Printf("Error unmarshaling json from %s\n[%s]\n[%v]", url, resp.Data, err)
			}
		}
	} else {
		err = fmt.Errorf("error in rest operation: endpoint %s:%+v %d %s %s %s", url, err, resp.Status, resp.Error, resp.Message, resp.Data)
	}

	return
}

// Do the REST call and parse the result
func (rc *RestClient) restByService(operation string, endPoint string, input interface{}, out interface{}) (body []byte, err error) { // FIXME unify the response of all services
	var (
		data []byte
		resp HttpResp
		r    *http.Response
		req  *http.Request
	)

	if input != nil {
		if data, err = json.Marshal(input); err != nil {
			fmt.Printf("error marshalling input:%v\n%v\n", input, err)
			return
		}
	}

	url := fmt.Sprintf("https://%s:9999/%v", rc.address, endPoint) // FIXME move the port to json config file
	if req, err = http.NewRequest(operation, url, bytes.NewBuffer(data)); err == nil {
		req.Header.Add("Authorization", rc.token)
		req.Header.Add("Content-Type", "application/json")
		client := &http.Client{}
		if r, err = client.Do(req); err == nil {
			defer r.Body.Close()
			if body, err = ioutil.ReadAll(r.Body); err == nil {
				if strings.HasPrefix(strings.TrimPrefix(endPoint, "/"), "pccserver") { // FIXME all the services should have the same struct
					var (
						rg       pccRespGeneric
						dataJson []byte
					)
					if err = json.Unmarshal(body, &rg); err == nil {
						if dataJson, err = json.Marshal(rg.Data); err != nil {
							fmt.Printf("Unarshalling errror in response payload parsing:\n%v\n%v\n", err, rg.Data)
							return
						}
					} else {
						fmt.Printf("Unarshalling errror in response : %s\n%s\n%v\n%v\n", url, string(debug.Stack()), err, string(body))
						return
					}
					resp = HttpResp{
						Path:      rg.Path,
						Status:    rg.Status,
						RequestId: rg.RequestId,
						Message:   rg.Message,
						Error:     rg.Error,
						Data:      dataJson,
					}
				} else {
					resp = HttpResp{
						Status: r.StatusCode,
						Data:   body,
					}
				}

				err = unmarshalHttpResp(url, &resp, out)
			}
		}
	}

	return
}

// GET
func (rc *RestClient) Get(endPoint string, out interface{}) (err error) {
	_, err = rc.restByService("GET", endPoint, nil, out)
	return
}

// DELETE
func (rc *RestClient) Delete(endPoint string, out interface{}) (err error) {
	_, err = rc.restByService("DELETE", endPoint, nil, out)
	return
}

// POST
func (rc *RestClient) Post(endPoint string, data interface{}, out interface{}) (err error) {
	_, err = rc.restByService("POST", endPoint, data, out)
	return
}

// PUT
func (rc *RestClient) Put(endPoint string, data interface{}, out interface{}) (err error) {
	_, err = rc.restByService("PUT", endPoint, data, out)
	return
}

// PUT file
func (rc *RestClient) PutFile(endPoint string, filePath string, fields map[string]string, out interface{}) (err error) { // FIXME all services should share the same structure
	var (
		file         *os.File
		writer       *multipart.Writer
		part         io.Writer
		req          *http.Request
		response     *http.Response
		responseBody []byte
	)
	if file, err = os.Open(filePath); err == nil {
		body := &bytes.Buffer{}
		writer = multipart.NewWriter(body)
		if fields != nil {
			for k, v := range fields {
				if err = writer.WriteField(k, v); err != nil {
					return
				}
			}
		}

		if part, err = writer.CreateFormFile("file", filepath.Base(filePath)); err == nil {
			if _, err = io.Copy(part, file); err == nil {
				if err = writer.Close(); err == nil {
					client := &http.Client{}
					url := fmt.Sprintf("https://%s:9999/%s", rc.address, endPoint)
					if req, err = http.NewRequest("POST", url, body); err == nil {
						req.Header.Add("Authorization", rc.token)
						req.Header.Set("Content-Type", writer.FormDataContentType())
						response, err = client.Do(req)
						defer response.Body.Close()
						if err == nil && out != nil {
							if responseBody, err = ioutil.ReadAll(response.Body); err == nil {
								var uploadGeneric struct {
									Body interface{} `json:"body"`
								}

								if err = json.Unmarshal(responseBody, &uploadGeneric); err == nil {
									if responseBody, err = json.Marshal(uploadGeneric.Body); err != nil {
										fmt.Printf("Marshalling Error:\n%v\n", err)
										return
									}
								} else {
									fmt.Printf("Unmarshalling Error: %s\n%v\n", string(debug.Stack()), string(responseBody))
									return
								}

								resp := HttpResp{Status: response.StatusCode, Data: responseBody}
								err = unmarshalHttpResp(url, &resp, out)
							}
						}
					}
				}
			}
		}
	}
	return
}

// GET a file
func (rc *RestClient) GetFile(endPoint string) (content string, err error) {
	var body []byte
	if body, err = rc.restByService("GET", endPoint, nil, nil); err == nil {
		content = string(body)
	}
	return
}
