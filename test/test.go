package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"

	"github.com/denisbrodbeck/machineid"
)

// 下面是一个生成私匙和公匙的方法
// privateKey, publicKey分别是私匙和公匙的文件可写流，私匙和公匙分别写入到这二个文件中
// bits为生成私匙的长度
func GenRsaKey(bits int) error {
	// 生成私钥文件
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "私钥",
		Bytes: derStream,
	}
	file, err := os.Create("private.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	// 生成公钥文件
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "公钥",
		Bytes: derPkix,
	}
	file, err = os.Create("public.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil

}

// 加密
func RsaEncrypt(publicKey, origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(privateKey, ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func main() {
	id, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)

	// var bits int
	// flag.IntVar(&bits, "b", 2048, "密钥长度，默认为2048位")
	// if err := GenRsaKey(2048); err != nil {
	// 	log.Fatal("密钥文件生成失败！")
	// }
	// log.Println("密钥文件生成成功！")

	publicKey, err := ioutil.ReadFile("public.pem")
	if err != nil {
		os.Exit(-1)
	}
	privateKey, err := ioutil.ReadFile("private.pem")
	if err != nil {
		os.Exit(-1)
	}

	fmt.Print("======================\n")
	data, err := RsaEncrypt(publicKey, []byte(id))
	if err != nil {
		panic(err)
	}
	origData, err := RsaDecrypt(privateKey, data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))

}
