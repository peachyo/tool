package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"flag"
	"strings"
	"sync"
)

type TestConfig struct {
	TSToken string `json : "tstoken"`
	TSURL   string `json : "tsurl"`
	Tests   []TestUnit `json : "tests"`
}

type TSTest struct {
	UID      string        `json:"uid"`
	AID      string        `json:"aid"`
	Token    string        `json:"request_token_template"`
	Register string        `json:"regsiter_template"`
	Bind     string        `json:"bind_template"`
	Login	 string	       `json:"login_template"`
}

type TestUnit struct {
	Test TSTest 	`json:"test"`
	Count int 	`json:"count"`
}

type result struct{
	device_id string
	device_id_logical string
	err error
	key_id string
}
var tsurl string
var tstoken string


func main() {
	var config = flag.String("testconfig", "config/testconfig.json", "test config file path")
	flag.Usage = func() {
		fmt.Println("Usage: devicetool -testconfig [testconfig.json]")
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("config:", *config)


	//read ts config file
	tsconfig := getConfig(*config)
	tsurl = tsconfig.TSURL
	tstoken = tsconfig.TSToken
	fmt.Println("TS token:", tstoken)
	fmt.Println("TS url:", tsurl)
	c := runtests(tsconfig.Tests)
	for r:= range c{
		fmt.Println("Result: ", r)
	}

}

func runtests(tests []TestUnit) <-chan result {

	c := make(chan result)

	go func() {
		var wg = sync.WaitGroup{}
		for _,test := range(tests){

			for i:=0; i<test.Count; i++{
				wg.Add(1)
				go run(test.Test, &wg, c)
			}

		}
		go func(){
			wg.Wait()
			close(c)
		}()
	}()

	return c
}

func run(test TSTest, wg *sync.WaitGroup, c chan<-result) {
	defer wg.Done()
	uid := test.UID
	aid := test.AID

	did, _ := newUUID()
	did = strings.ToUpper(did)

	priv_pem, pub_pem := pems()
	pub_key := getPublicKey(pub_pem)

	token_template := getTemplate(test.Token)

	token, err:= tokenRequest(uid, aid, did, token_template)
	if err!= nil {
		c<-result{did, "", err, ""}
		return
	}

	register_template := getTemplate(test.Register)
	device_id, err := register(uid, aid, pub_key, token, did, register_template)
	if err!= nil {
		c<-result{did, "", err, ""}
        	return
	}

	if device_id == "" {
		c<-result{did, "", err, ""}
		return
	}
	
	bind_template := getTemplate(test.Bind)
	key_id, err := bind(uid, aid, pub_key, device_id, bind_template)
		if err!= nil {
        		c<-result{did, device_id, err, ""}
                	return
        	}
	fmt.Println("Bind device id:", key_id)

	if key_id == "" {
			if err!= nil {
                		c<-result{did, device_id, err, ""}
                        	return
                	}
	}

	login_template := getTemplate(test.Login)
	//fmt.Println("Login....")
	login(uid, aid, priv_pem, key_id, device_id, login_template)
	c<-result{did, device_id, err, key_id}
}

func getConfig(configFile string) TestConfig {
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err.Error())
	}

	var tsconfig TestConfig
	if err := json.Unmarshal(raw, &tsconfig); err != nil {
		fmt.Println(err.Error())
	}

	return tsconfig
}

func getTemplate(templateFile string) string {
	raw, err := ioutil.ReadFile(templateFile)
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(raw)
}


