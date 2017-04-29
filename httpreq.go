package main

import (
	"fmt"
	"net/http"
	"bytes"
	"strings"
	"io/ioutil"
	"time"
)

func tokenRequest(uid, aid, did, token_template string) (string, error){
	url := tsurl +  "/api/v2/token/create?aid=" + aid + "&uid=" + uid

	didstr, err := getEncodedDeviceStr(did)
	if err!=nil{
		return "", err
	}
	datastr := strings.Replace(token_template, "$ENCODED_DEVICE_ID", didstr, 1)

	req,err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(datastr)))
	req.Header.Set("Authorization", tstoken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp,err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println("request token response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println("response body:", string(body))
	if err != nil {
		return "", err
	}
	return getString(string(body), "token"), nil

}

func register(uid, aid, pub_key, token, device_id, data_template string) (string, error) {
	url := tsurl +  "/api/v2/auth/device/register?uid=" + uid + "&aid=" + aid
	data := strings.Replace(strings.Replace((strings.Replace(data_template, "$PUB_PEM", pub_key, 1)),"$DEVICE_ID", device_id, 1),  "$TOKEN", token, 1)

	req,err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err!=nil{
		return "", err
	}
	req.Header.Set("Authorization", tstoken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp,err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println("register device response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}
	return getString(string(body), "device_id"), nil
}

func bind(uid, aid, pub_key, device_id, data_template string) (string, error){
	start := time.Now()
	url := tsurl +  "/api/v2/auth/bind?uid=" + uid + "&aid=" + aid
	data := strings.Replace((strings.Replace(data_template , "$PUB_PEM", pub_key, 1)),"$DEVICE_ID",device_id, 1)

	req,err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Authorization", tstoken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println("bind device response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)

	if(err!=nil){
		return "", err
	}
	timeTrack(start, "bind device")
	return getString(string(body), "device_id"), nil
}

func login(uid, aid, priv_pem, key_id, device_id, data_template string) error {
	urlstr := "/api/v2/auth/login?uid=" + uid + "&did=" + key_id + "&aid=" + aid
	data := strings.Replace(data_template, "$DEVICE_ID", device_id, 1)
	sig, err:= contentSig(priv_pem, urlstr, data, key_id)
	if err!=nil{
		return err
	}

	url := tsurl + urlstr

	req,err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Authorization", tstoken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Sigature", sig)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("login response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response body:", string(body))
	if(err!=nil){
		return err
	}
	return nil
}


