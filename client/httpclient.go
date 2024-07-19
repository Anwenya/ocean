package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = NewHTTPClient()

type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient
// 通用http请求客户端
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				MaxConnsPerHost:       20,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// Request http请求
type Request struct {
	request *http.Request
	isBad   bool
}

// NewRequest http请求构造函数
func NewRequest(method, host, path string) *Request {
	request, err := http.NewRequest(method, host+path, nil)
	if err != nil {
		fmt.Printf("new request fail:%v\n", err)
		return &Request{
			request: request,
			isBad:   true,
		}
	}
	return &Request{
		request: request,
		isBad:   false,
	}
}

// SetBad
// 组装请求过程中出现异常
func (r *Request) SetBad() {
	r.isBad = true
}

// Headers 设置请求头
func (r *Request) Headers(headers map[string]string) *Request {
	for k, v := range headers {
		r.request.Header.Set(k, v)
	}
	return r
}

// Json 添加json参数
// 不对参数做强校验,使用时自己留心
func (r *Request) Json(param interface{}) *Request {
	jsonBytes, err := json.Marshal(param)
	if err != nil {
		fmt.Printf("json marshal fail:%v\n", err)
		r.SetBad()
		return r
	}
	jsonBuffer := bytes.NewBuffer(jsonBytes)
	rc := io.NopCloser(jsonBuffer)
	r.request.Body = rc
	r.request.ContentLength = int64(jsonBuffer.Len())
	r.request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(jsonBuffer.Bytes())
		return io.NopCloser(r), nil
	}
	r.request.Header.Set("Content-Type", "application/json; charset=utf-8")
	return r
}

// Param 添加url参数
// 不对参数做强校验,使用时自己留心
func (r *Request) Param(params url.Values) *Request {
	r.request.URL.RawQuery = params.Encode()
	return r
}

// Form 添加表单参数
// 不对参数做强校验,使用时自己留心
func (r *Request) Form(params url.Values) *Request {
	formReader := strings.NewReader(params.Encode())
	rc := io.NopCloser(formReader)
	r.request.Body = rc
	r.request.ContentLength = int64(formReader.Len())
	snapshot := *formReader
	r.request.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return io.NopCloser(&r), nil
	}
	r.request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	return r
}

// Do 发送请求
func (r *Request) Do() (*Response, error) {
	if r.isBad {
		fmt.Println("bad request, request fail")
		return nil, errors.New("bad request")
	}
	r.request.Header.Set("Accept-Charset", "utf-8")
	resp, err := client.client.Do(r.request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

func Get(host, path string) *Request {
	return NewRequest(http.MethodGet, host, path)
}

func Post(host, path string) *Request {
	return NewRequest(http.MethodPost, host, path)
}

func Delete(host, path string) *Request {
	return NewRequest(http.MethodDelete, host, path)
}

func Put(host, path string) *Request {
	return NewRequest(http.MethodPut, host, path)
}

// Response http响应
type Response struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// JsonUnmarshal json反序列化
func (r *Response) JsonUnmarshal(data interface{}) error {
	return json.Unmarshal(r.Body, data)
}

// IsOK 判断状态码是不是200 - 300
// 有的接口不一定以状态码来区分成功/失败 请以实际情况判断是否使用该方法
func (r *Response) IsOK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
