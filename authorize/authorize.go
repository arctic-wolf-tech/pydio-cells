/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package authorize

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/denisbrodbeck/machineid"
)

var privateKeyStr string = `-----BEGIN 私钥-----
MIIEowIBAAKCAQEAvbIph0tEan8amCuLhdJDDUS2XGGF5cGMaKhitfsfUNtpXNbi
MLT31xZUydDcRFfKx8a8+6MDxD1udOyfiUQINsKLfeDkKYRlHXdxWsxo+/4d+IzC
jGWC8MM1ebzLRd3/V4PEyaYjWFgPOiN3M4xPxzhTzn4T3n9AcRsZTEax2ohhCYGN
AwejrdYE2WXvogB8r+WmgmMcTXPeqbwRg1XKLWnh1M8YQeUTETwCUV3xDEbGM9h5
/pUfNJbm702SDzfBLRNso83yaZNMAGFdfEGFKPrde1VGg6hu7YY9oIRAU/qJRZbQ
evMEX7FUBqSUaR+IFJVw+5q0MQNR710EElYYRQIDAQABAoIBAESJ3cVbZZHQ8Mvw
V833JXDi1bzVI6ra3p9lz5yO6katsAjyPvF4QV/+Wo48n4k16zd5UAjfYloCFCm8
4PuYkBsw+XN20RlLE7ms0VEMMBZ0P2Hxgc12U/Qno+ejVhKdXkfBfVWaaITf9Eh+
TfBbDuwdJvKhzQ4EDkWPk/liRZp+MVZ5QO/DNowwjDC9x1n2nJeog6LoQF1jPLAT
w50s+lvBdEBDakfTqnJxie/TqCTEKOwk2nuk8VxNltox7++o5dFqgriuhsHHPB4o
Rxf8/4eDhnf+jJi2N+ZIU6cufRYiflHVWeAdrxyVbATyYf7sVzkbbgo3YrTLK3DT
CRC9HJUCgYEA/D1TorxjzMCuOgBaEZi8GIrktYm6Wj/GCdoVHWkuVZjrqIBf8apI
lHgC3La+6HYsgeGxNU58RtmazyS8+PLmlyh+Z2fvZT8n7tBedk320eBcV/CjUTnj
hK3ZyYPXtp/J6c6tz+vTP0rN0hPtUDCPYlaMdLsvD6t+CYIHPSMQsHMCgYEAwIYj
N+aRyAjwelKIbEWjPkUO1KFkUdumFvC1SCOIKs6NNVJnv/2dvuRXDrv07evSZZ3H
116pd6gibZH4WjcLVv1H7H7f+UkwN3k+GlN4g0w4AmgDLUQd7auksRfGYn7XQ0t+
5IrC2DQueiu669tuYdvjhdFqphvf0F43PIf//mcCgYBTHnJdAe9xHV1MR6lmewog
nERZfhUmgDVmMbMbifl2w3mEgSkcnZxlMFbhHGc0exyXgCPBCSfywOo+sECFWKWb
0gA1Ww6MMo+aJpe6LF7VMjW71NQ8g/LxWciWmxeOWoFSxoSIK5HlHWVNgLuG3Tmi
khqerMAJTd2ujGaOQuvQ+QKBgQCEJA1MMw9gUvJrovZMCkgPV2rkepnWrYIEQNbM
Wsb9SqQVMyhO2I5LFYLDdDKakr/oSzF9G1YJ8PcgaY4iraE05cdWBYdJHPjhOnBr
tVsEE25mCGoVyakZFjSF8KGTvSeW4tyHlM5Dgx1bcWRsukG7HSe/E4u102/9Ho2f
GGGWhQKBgGHxGQC4ZCqIBZWJYb7vhLjamsUZWz2IwIFIyDQcTn+odBh4qNeeqb8+
aoXMJSqzXCShKAlnXoc9nWpx2uWp/lmk3BoxN44Pmi+h2Nt9vDD5/VVzNsvwS1rQ
Vj7yCgji3bs/txkUxNjcHsFPHRpJbmgV74Ug0GvJUFuAu8nTIus9
-----END 私钥-----
`

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

func Authorize(exePath string) {
	parrentDir := path.Dir(exePath)

	fmt.Println("The authorize file root dir: " + parrentDir)
	id, err := machineid.ID()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("The machine id is " + id)

	// 获取私钥
	privateKey := []byte(privateKeyStr)

	// 读取证书信息
	fmt.Printf("License path : %s\n", filepath.Join(parrentDir, "pydio.license"))
	license, err := ioutil.ReadFile(filepath.Join(parrentDir, "pydio.license"))
	if err != nil {
		fmt.Printf("Error happened Reade license: %s\n", err)
		os.Exit(1)
	}

	// 解密证书信息
	id_decrypt, err := RsaDecrypt(privateKey, license)
	if err != nil {
		fmt.Printf("Error happened RsaDecrypt license: %s\n", err)
		os.Exit(1)
	}

	if string(id_decrypt) == id {
		fmt.Println("license authorize pass, machine id is: " + id + ", machine id decrypted is " + string(id_decrypt))
	} else {
		fmt.Println("license authorize fail, machine id is: " + id + ", machine id decrypted is " + string(id_decrypt))
		os.Exit(1)
	}
}
