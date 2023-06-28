package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/clbanning/mxj"
)

// Signature :
type Signature struct {
	Sign      string
	NonceStr  string
	Timestamp string
	Data      string
}

// Sign Type
const (
	MD4        = "md4"
	MD5        = "md5"
	MD5SHA1    = "md5-sha1"
	SHA1       = "sha1"
	SHA224     = "sha224"
	SHA256     = "sha256"
	SHA384     = "sha384"
	SHA3224    = "sha3-224"
	SHA3256    = "sha3-256"
	SHA3384    = "sha3-384"
	SHA3512    = "sha3-512"
	SHA512     = "sha512"
	SHA512224  = "sha512-224"
	SHA512256  = "sha512-256"
	BLAKE2B256 = "blake2b-256"
	BLAKE2B384 = "blake2b-384"
	BLAKE2B512 = "blake2b-512"
	BLAKE2S256 = "blake2s-256"
)

var signTypeMapper = map[string]crypto.Hash{
	MD4:        crypto.MD4,
	MD5:        crypto.MD5,
	MD5SHA1:    crypto.MD5SHA1,
	SHA1:       crypto.SHA1,
	SHA224:     crypto.SHA224,
	SHA256:     crypto.SHA256,
	SHA384:     crypto.SHA384,
	SHA3224:    crypto.SHA3_224,
	SHA3256:    crypto.SHA3_256,
	SHA3384:    crypto.SHA3_384,
	SHA3512:    crypto.SHA3_512,
	SHA512:     crypto.SHA512,
	SHA512224:  crypto.SHA512_224,
	SHA512256:  crypto.SHA512_256,
	BLAKE2B256: crypto.BLAKE2b_256,
	BLAKE2B384: crypto.BLAKE2b_384,
	BLAKE2B512: crypto.BLAKE2b_512,
	BLAKE2S256: crypto.BLAKE2s_256,
}

// Generate :
func Generate(timestamp, nonceStr, method, requestURL, signType, body string, privateKey []byte) (*Signature, error) {
	method = strings.ToLower(method)
	h, isExist := signTypeMapper[signType]
	if !isExist {
		return nil, errors.New("sign type not found")
	}

	data, err := formData(timestamp, nonceStr, method, requestURL, signType, body)
	if err != nil {
		return nil, err
	}

	signature, err := rsa2Hash(h, []byte(data), privateKey)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Sign:      fmt.Sprintf("%s %s", signType, signature),
		NonceStr:  nonceStr,
		Timestamp: timestamp,
		Data:      data,
	}, nil
}

// Validate :
func Validate(timestamp, nonceStr, method, requestURL, signature, body string, publicKey []byte) bool {
	method = strings.ToLower(method)
	signArr := strings.Split(signature, " ")

	if len(signArr) != 2 {
		return false
	}

	signType := signArr[0]
	signature = signArr[1]

	h, isExist := signTypeMapper[signType]
	if !isExist {
		return false
	}

	data, err := formData(timestamp, nonceStr, method, requestURL, signType, body)
	if err != nil {
		return false
	}

	if err := verifyPKCS1v15(h, publicKey, []byte(data), signature); err != nil {
		return false
	}

	return true
}

// SignForm :
type SignForm struct {
	Data       string `json:"data,omitempty"`
	Method     string `json:"method"`
	NonceStr   string `json:"nonceStr"`
	RequestURL string `json:"requestUrl,omitempty"`
	SignType   string `json:"signType"`
	Timestamp  string `json:"timestamp"`
}

func formData(timestamp, nonceStr, method, requestURL, signType, body string) (string, error) {
	s := SignForm{
		Data:       body,
		Timestamp:  timestamp,
		NonceStr:   nonceStr,
		Method:     method,
		RequestURL: requestURL,
		SignType:   signType,
	}

	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	data, err := sortJSONByKey(jsonData)
	if err != nil {
		return "", err
	}

	return data, nil
}

func rsa2Hash(h crypto.Hash, data, privateKey []byte) (string, error) {
	signature, err := signPKCS1v15(h, data, privateKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func signPKCS1v15(hash crypto.Hash, src, key []byte) ([]byte, error) {
	var h = hash.New()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}

	hashed := h.Sum(nil)

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}

	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.SignPKCS1v15(rand.Reader, pk, hash, hashed)
}

func verifyPKCS1v15(ch crypto.Hash, publicKey []byte, data []byte, signature string) error {
	var hashed []byte

	switch ch {
	case crypto.SHA256:
		var h = sha256.New()
		h.Write(data)
		hashed = h.Sum(nil)
	}

	block, _ := pem.Decode(publicKey)
	if block == nil {
		return errors.New("public key error")
	}

	var pubInterface interface{}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	var pub = pubInterface.(*rsa.PublicKey)

	signByte, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, signByte)
}

func sortJSONByKey(b []byte) (string, error) {
	m, err := mxj.NewMapJson(b)
	if err != nil {
		return "", err
	}

	m = m.Old()
	arr := make([]string, 0)
	for k, v := range m {
		if strings.TrimSpace(strings.ToLower(k)) == "sign" {
			continue
		}

		str := ""
		switch vi := v.(type) {
		case map[string]interface{}:
			continue
		case string:
			if vi == "" || vi == "{}" {
				continue
			}
			str = vi
		default:
			str = fmt.Sprintf("%v", vi)
		}
		arr = append(arr, fmt.Sprintf("%s=%s", k, str))
	}

	sort.Sort(sort.StringSlice(arr))

	return strings.Join(arr, "&"), nil
}
