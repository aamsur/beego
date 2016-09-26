// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/aamsur/beego/session"
)

// BeegoInput operates the http request header, data, cookie and body.
// it also contains router params and current session.
type BeegoInput struct {
	CruSession    session.SessionStore
	Params        map[string]string
	Data          map[interface{}]interface{} // store some values in this context when calling context in filter or controller.
	Request       *http.Request
	RequestBody   []byte
	RunController reflect.Type
	RunMethod     string
}

// NewInput return BeegoInput generated by http.Request.
func NewInput(req *http.Request) *BeegoInput {
	return &BeegoInput{
		Params:  make(map[string]string),
		Data:    make(map[interface{}]interface{}),
		Request: req,
	}
}

// Protocol returns request protocol name, such as HTTP/1.1 .
func (input *BeegoInput) Protocol() string {
	return input.Request.Proto
}

// Uri returns full request url with query string, fragment.
func (input *BeegoInput) Uri() string {
	return input.Request.RequestURI
}

// Url returns request url path (without query string, fragment).
func (input *BeegoInput) Url() string {
	return input.Request.URL.Path
}

// Site returns base site url as scheme://domain type.
func (input *BeegoInput) Site() string {
	return input.Scheme() + "://" + input.Domain()
}

// Scheme returns request scheme as "http" or "https".
func (input *BeegoInput) Scheme() string {
	if input.Request.URL.Scheme != "" {
		return input.Request.URL.Scheme
	}
	if input.Request.TLS == nil {
		return "http"
	}
	return "https"
}

// Domain returns host name.
// Alias of Host method.
func (input *BeegoInput) Domain() string {
	return input.Host()
}

// Host returns host name.
// if no host info in request, return localhost.
func (input *BeegoInput) Host() string {
	if input.Request.Host != "" {
		hostParts := strings.Split(input.Request.Host, ":")
		if len(hostParts) > 0 {
			return hostParts[0]
		}
		return input.Request.Host
	}
	return "localhost"
}

// Method returns http request method.
func (input *BeegoInput) Method() string {
	return input.Request.Method
}

// Is returns boolean of this request is on given method, such as Is("POST").
func (input *BeegoInput) Is(method string) bool {
	return input.Method() == method
}

// Is this a GET method request?
func (input *BeegoInput) IsGet() bool {
	return input.Is("GET")
}

// Is this a POST method request?
func (input *BeegoInput) IsPost() bool {
	return input.Is("POST")
}

// Is this a Head method request?
func (input *BeegoInput) IsHead() bool {
	return input.Is("HEAD")
}

// Is this a OPTIONS method request?
func (input *BeegoInput) IsOptions() bool {
	return input.Is("OPTIONS")
}

// Is this a PUT method request?
func (input *BeegoInput) IsPut() bool {
	return input.Is("PUT")
}

// Is this a DELETE method request?
func (input *BeegoInput) IsDelete() bool {
	return input.Is("DELETE")
}

// Is this a PATCH method request?
func (input *BeegoInput) IsPatch() bool {
	return input.Is("PATCH")
}

// IsAjax returns boolean of this request is generated by ajax.
func (input *BeegoInput) IsAjax() bool {
	return input.Header("X-Requested-With") == "XMLHttpRequest"
}

// IsSecure returns boolean of this request is in https.
func (input *BeegoInput) IsSecure() bool {
	return input.Scheme() == "https"
}

// IsWebsocket returns boolean of this request is in webSocket.
func (input *BeegoInput) IsWebsocket() bool {
	return input.Header("Upgrade") == "websocket"
}

// IsUpload returns boolean of whether file uploads in this request or not..
func (input *BeegoInput) IsUpload() bool {
	return strings.Contains(input.Header("Content-Type"), "multipart/form-data")
}

// IP returns request client ip.
// if in proxy, return first proxy id.
// if error, return 127.0.0.1.
func (input *BeegoInput) IP() string {
	ips := input.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		rip := strings.Split(ips[0], ":")
		return rip[0]
	}
	ip := strings.Split(input.Request.RemoteAddr, ":")
	if len(ip) > 0 {
		if ip[0] != "[" {
			return ip[0]
		}
	}
	return "127.0.0.1"
}

// Proxy returns proxy client ips slice.
func (input *BeegoInput) Proxy() []string {
	if ips := input.Header("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

// Referer returns http referer header.
func (input *BeegoInput) Referer() string {
	return input.Header("Referer")
}

// Refer returns http referer header.
func (input *BeegoInput) Refer() string {
	return input.Referer()
}

// SubDomains returns sub domain string.
// if aa.bb.domain.com, returns aa.bb .
func (input *BeegoInput) SubDomains() string {
	parts := strings.Split(input.Host(), ".")
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-2], ".")
	}
	return ""
}

// Port returns request client port.
// when error or empty, return 80.
func (input *BeegoInput) Port() int {
	parts := strings.Split(input.Request.Host, ":")
	if len(parts) == 2 {
		port, _ := strconv.Atoi(parts[1])
		return port
	}
	return 80
}

// UserAgent returns request client user agent string.
func (input *BeegoInput) UserAgent() string {
	return input.Header("User-Agent")
}

// Param returns router param by a given key.
func (input *BeegoInput) Param(key string) string {
	if v, ok := input.Params[key]; ok {
		return v
	}
	return ""
}

// Query returns input data item string by a given string.
func (input *BeegoInput) Query(key string) string {
	if val := input.Param(key); val != "" {
		return val
	}
	if input.Request.Form == nil {
		input.Request.ParseForm()
	}
	return input.Request.Form.Get(key)
}

// Header returns request header item string by a given string.
// if non-existed, return empty string.
func (input *BeegoInput) Header(key string) string {
	return input.Request.Header.Get(key)
}

// Cookie returns request cookie item string by a given key.
// if non-existed, return empty string.
func (input *BeegoInput) Cookie(key string) string {
	ck, err := input.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return ck.Value
}

// Session returns current session item value by a given key.
// if non-existed, return empty string.
func (input *BeegoInput) Session(key interface{}) interface{} {
	return input.CruSession.Get(key)
}

// CopyBody returns the raw request body data as bytes.
func (input *BeegoInput) CopyBody() []byte {
	requestbody, _ := ioutil.ReadAll(input.Request.Body)
	input.Request.Body.Close()
	bf := bytes.NewBuffer(requestbody)
	input.Request.Body = ioutil.NopCloser(bf)
	input.RequestBody = requestbody
	return requestbody
}

// GetData returns the stored data in this context.
func (input *BeegoInput) GetData(key interface{}) interface{} {
	if v, ok := input.Data[key]; ok {
		return v
	}
	return nil
}

// SetData stores data with given key in this context.
// This data are only available in this context.
func (input *BeegoInput) SetData(key, val interface{}) {
	input.Data[key] = val
}

// parseForm or parseMultiForm based on Content-type
func (input *BeegoInput) ParseFormOrMulitForm(maxMemory int64) error {
	// Parse the body depending on the content type.
	if strings.Contains(input.Header("Content-Type"), "multipart/form-data") {
		if err := input.Request.ParseMultipartForm(maxMemory); err != nil {
			return errors.New("Error parsing request body:" + err.Error())
		}
	} else if err := input.Request.ParseForm(); err != nil {
		return errors.New("Error parsing request body:" + err.Error())
	}
	return nil
}

// Bind data from request.Form[key] to dest
// like /?id=123&isok=true&ft=1.2&ol[0]=1&ol[1]=2&ul[]=str&ul[]=array&user.Name=astaxie
// var id int  beegoInput.Bind(&id, "id")  id ==123
// var isok bool  beegoInput.Bind(&isok, "isok")  id ==true
// var ft float64  beegoInput.Bind(&ft, "ft")  ft ==1.2
// ol := make([]int, 0, 2)  beegoInput.Bind(&ol, "ol")  ol ==[1 2]
// ul := make([]string, 0, 2)  beegoInput.Bind(&ul, "ul")  ul ==[str array]
// user struct{Name}  beegoInput.Bind(&user, "user")  user == {Name:"astaxie"}
func (input *BeegoInput) Bind(dest interface{}, key string) error {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return errors.New("beego: non-pointer passed to Bind: " + key)
	}
	value = value.Elem()
	if !value.CanSet() {
		return errors.New("beego: non-settable variable passed to Bind: " + key)
	}
	rv := input.bind(key, value.Type())
	if !rv.IsValid() {
		return errors.New("beego: reflect value is empty")
	}
	value.Set(rv)
	return nil
}

func (input *BeegoInput) bind(key string, typ reflect.Type) reflect.Value {
	rv := reflect.Zero(reflect.TypeOf(0))
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindInt(val, typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindUint(val, typ)
	case reflect.Float32, reflect.Float64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindFloat(val, typ)
	case reflect.String:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindString(val, typ)
	case reflect.Bool:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindBool(val, typ)
	case reflect.Slice:
		rv = input.bindSlice(&input.Request.Form, key, typ)
	case reflect.Struct:
		rv = input.bindStruct(&input.Request.Form, key, typ)
	case reflect.Ptr:
		rv = input.bindPoint(key, typ)
	case reflect.Map:
		rv = input.bindMap(&input.Request.Form, key, typ)
	}
	return rv
}

func (input *BeegoInput) bindValue(val string, typ reflect.Type) reflect.Value {
	rv := reflect.Zero(reflect.TypeOf(0))
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv = input.bindInt(val, typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv = input.bindUint(val, typ)
	case reflect.Float32, reflect.Float64:
		rv = input.bindFloat(val, typ)
	case reflect.String:
		rv = input.bindString(val, typ)
	case reflect.Bool:
		rv = input.bindBool(val, typ)
	case reflect.Slice:
		rv = input.bindSlice(&url.Values{"": {val}}, "", typ)
	case reflect.Struct:
		rv = input.bindStruct(&url.Values{"": {val}}, "", typ)
	case reflect.Ptr:
		rv = input.bindPoint(val, typ)
	case reflect.Map:
		rv = input.bindMap(&url.Values{"": {val}}, "", typ)
	}
	return rv
}

func (input *BeegoInput) bindInt(val string, typ reflect.Type) reflect.Value {
	intValue, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetInt(intValue)
	return pValue.Elem()
}

func (input *BeegoInput) bindUint(val string, typ reflect.Type) reflect.Value {
	uintValue, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetUint(uintValue)
	return pValue.Elem()
}

func (input *BeegoInput) bindFloat(val string, typ reflect.Type) reflect.Value {
	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetFloat(floatValue)
	return pValue.Elem()
}

func (input *BeegoInput) bindString(val string, typ reflect.Type) reflect.Value {
	return reflect.ValueOf(val)
}

func (input *BeegoInput) bindBool(val string, typ reflect.Type) reflect.Value {
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "true", "on", "1":
		return reflect.ValueOf(true)
	}
	return reflect.ValueOf(false)
}

type sliceValue struct {
	index int           // Index extracted from brackets.  If -1, no index was provided.
	value reflect.Value // the bound value for this slice element.
}

func (input *BeegoInput) bindSlice(params *url.Values, key string, typ reflect.Type) reflect.Value {
	maxIndex := -1
	numNoIndex := 0
	sliceValues := []sliceValue{}
	for reqKey, vals := range *params {
		if !strings.HasPrefix(reqKey, key+"[") {
			continue
		}
		// Extract the index, and the index where a sub-key starts. (e.g. field[0].subkey)
		index := -1
		leftBracket, rightBracket := len(key), strings.Index(reqKey[len(key):], "]")+len(key)
		if rightBracket > leftBracket+1 {
			index, _ = strconv.Atoi(reqKey[leftBracket+1 : rightBracket])
		}
		subKeyIndex := rightBracket + 1

		// Handle the indexed case.
		if index > -1 {
			if index > maxIndex {
				maxIndex = index
			}
			sliceValues = append(sliceValues, sliceValue{
				index: index,
				value: input.bind(reqKey[:subKeyIndex], typ.Elem()),
			})
			continue
		}

		// It's an un-indexed element.  (e.g. element[])
		numNoIndex += len(vals)
		for _, val := range vals {
			// Unindexed values can only be direct-bound.
			sliceValues = append(sliceValues, sliceValue{
				index: -1,
				value: input.bindValue(val, typ.Elem()),
			})
		}
	}
	resultArray := reflect.MakeSlice(typ, maxIndex+1, maxIndex+1+numNoIndex)
	for _, sv := range sliceValues {
		if sv.index != -1 {
			resultArray.Index(sv.index).Set(sv.value)
		} else {
			resultArray = reflect.Append(resultArray, sv.value)
		}
	}
	return resultArray
}

func (input *BeegoInput) bindStruct(params *url.Values, key string, typ reflect.Type) reflect.Value {
	result := reflect.New(typ).Elem()
	fieldValues := make(map[string]reflect.Value)
	for reqKey, val := range *params {
		if !strings.HasPrefix(reqKey, key+".") {
			continue
		}

		fieldName := reqKey[len(key)+1:]

		if _, ok := fieldValues[fieldName]; !ok {
			// Time to bind this field.  Get it and make sure we can set it.
			fieldValue := result.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				continue
			}
			if !fieldValue.CanSet() {
				continue
			}
			boundVal := input.bindValue(val[0], fieldValue.Type())
			fieldValue.Set(boundVal)
			fieldValues[fieldName] = boundVal
		}
	}

	return result
}

func (input *BeegoInput) bindPoint(key string, typ reflect.Type) reflect.Value {
	return input.bind(key, typ.Elem()).Addr()
}

func (input *BeegoInput) bindMap(params *url.Values, key string, typ reflect.Type) reflect.Value {
	var (
		result    = reflect.MakeMap(typ)
		keyType   = typ.Key()
		valueType = typ.Elem()
	)
	for paramName, values := range *params {
		if !strings.HasPrefix(paramName, key+"[") || paramName[len(paramName)-1] != ']' {
			continue
		}

		key := paramName[len(key)+1 : len(paramName)-1]
		result.SetMapIndex(input.bindValue(key, keyType), input.bindValue(values[0], valueType))
	}
	return result
}
