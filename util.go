package main

import(
	"fmt"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto/rand"
	"os"
	"crypto/sha256"
	"crypto"
	"encoding/base64"
	"io"
	"strings"
	"io/ioutil"
	"time"
	"errors"
)

func getPublicKey(pub_pem string)string{
	s1:=strings.TrimPrefix(pub_pem, "-----BEGIN PUBLIC KEY-----")
	s2:=strings.Replace(s1, "-----END PUBLIC KEY-----", "", 1)
	s3:=strings.Replace(s2, "\n", "", -1)
	return strings.TrimSpace(s3)
}
func pems()(string, string) {

	// Generate RSA Keys
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	priv_der := x509.MarshalPKCS1PrivateKey(priv);
	
	priv_blk := pem.Block {
		Type: "RSA PRIVATE KEY",
		Headers: nil,
		Bytes: priv_der,
	};

	priv_pem := string(pem.EncodeToMemory(&priv_blk));


	pub := priv.PublicKey;
	pub_der, _ := x509.MarshalPKIXPublicKey(&pub);
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}

	pub_blk := pem.Block {
		Type: "PUBLIC KEY",
		Headers: nil,
		Bytes: pub_der,
	}
	pub_pem := string(pem.EncodeToMemory(&pub_blk));

	return priv_pem, pub_pem
}

func contentSig(priv_pem, url, data, key_id string) (string, error){

	block, _ := pem.Decode([]byte(priv_pem))
	if block == nil  {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	message := []byte(url+data)

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example
	PSSmessage := message
	pssh := sha256.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hashed, &opts)

	if err != nil {
		return "", err
	}

	sig :="data:" + base64.StdEncoding.EncodeToString(signature) + ";key-id:" + key_id

	return sig, nil
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func getString(resp, field string) string {
	index:= strings.LastIndex(resp, field)
	if index<0 {
		return ""
	}

	substr := resp[index:]
	z :=strings.Split(substr, ":")
	x:= strings.Replace(z[1], "}", " ", -1)
	r:=strings.TrimSpace(x)
	deviceId:=strings.Trim(r, `"`)
	return deviceId
}


func getEncodedDeviceStr(did string) (string, error){
	raw, err := ioutil.ReadFile("config/device_id.json")
	if err != nil {
		return "", err
	}
	text := string(raw)
	deviceData := strings.Replace(text, "$DEVICE_ID", did, 1)
	return base64.StdEncoding.EncodeToString([]byte(deviceData)), nil
}

func timeTrack(start time.Time, name string){
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}




