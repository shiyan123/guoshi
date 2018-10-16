package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HttpPost(postUrl string, q url.Values) ([]byte, error) {
	resp, err := http.PostForm(postUrl, q)
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s\n", postUrl, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[http] read err %s, %s\n", postUrl, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http] status err %s, %d, body: %s\n", postUrl, resp.StatusCode, string(b))
	}
	return b, nil
}

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func HttpGetWithParams(getUrl string, param url.Values) ([]byte, error) {
	if len(param) != 0 {
		params := ""
		for k, v := range param {
			p := fmt.Sprintf("%s=%s&", k, url.QueryEscape(v[0]))
			params = params + p
		}
		params = params[:len(params)-1]
		getUrl = getUrl + "?" + params
	}

	resp, err := http.Get(getUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http get] status err %s, %d\n", getUrl, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func HttpPostJson(url string, jsonObj interface{}) ([]byte, error) {
	requestBytes, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(requestBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func HttpPostWithIOReader(url string, body io.Reader) ([]byte, error) {
	resp, err := http.Post(url, "application/json; charset=utf-8", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func HttpDelete(url string) ([]byte, error) {
	client := http.DefaultClient
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http delete] status err %s, %d\n", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
