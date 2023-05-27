package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if getEnv("IMPLICAUZANT_SECRET_KEY", "") == "" {
		log.Fatal("IMPLICAUZANT_SECRET_KEY is not set")
		return
	}
	if getEnv("IMPLICAUZANT_SALT", "") == "" {
		log.Fatal("IMPLICAUZANT_SALT is not set")
		return
	}
	http.HandleFunc("/authorize", authorize)
	http.ListenAndServe(":8090", nil)
}

func authorize(w http.ResponseWriter, req *http.Request) {
	param := getParam(req.URL.Query())
	err := validateParam(param)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	hash := getHash(param)

	if req.Method == "GET" {
		tpl, err := template.ParseFiles("./authorize.html")
		if err != nil {
			log.Fatal(err)
		}
		if err := tpl.Execute(w, map[string]interface{}{
			"client_id": param.Client_id,
			"hash":      hash,
		}); err != nil {
			log.Fatal(err)
		}
	} else if req.Method == "POST" {
		req.ParseForm()
		name := req.Form.Get("name")
		password := req.Form.Get("password")
		input_hash := req.Form.Get("hash")
		if name == "" || password == "" {
			fmt.Fprintf(w, "required parameter is missing")
			return
		}
		if hash != input_hash {
			fmt.Fprintf(w, "verify failed. %s != %s", hash, input_hash)
			return
		}
		res, err := getIdToken(param, name, password)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "internal error")
			return
		}
		uri := param.Redirect_uri + "#" + res

		w.Header().Set("Location", uri)
		w.WriteHeader(302)
	}
}

type Param struct {
	Scope         string
	Response_type string
	Client_id     string
	Redirect_uri  string
	State         string
	Nonce         string
}

func getParam(query url.Values) Param {
	return Param{
		query.Get("scope"),
		query.Get("response_type"),
		query.Get("client_id"),
		query.Get("redirect_uri"),
		query.Get("state"),
		query.Get("nonce")}
}

func validateParam(param Param) error {
	if strings.Index(param.Scope, "openid") < 0 {
		return fmt.Errorf("scope must include openid")
	}
	if param.Response_type != "id_token" {
		return fmt.Errorf("response_type must be id_token")
	}
	if param.Client_id == "" {
		return fmt.Errorf("client_id must be set")
	}
	if param.Redirect_uri == "" {
		return fmt.Errorf("redirect_uri must be set")
	}
	if strings.Index(param.Redirect_uri, "https://") != 0 &&
		strings.Index(param.Redirect_uri, "http://") != 0 {
		return fmt.Errorf("redirect_uri must start with https://")
	}
	if param.State == "" {
		return fmt.Errorf("state must be set")
	}
	if param.Nonce == "" {
		return fmt.Errorf("nonce must be set")
	}
	return nil
}

func getHash(param Param) string {
	salt := getEnv("IMPLICAUZANT_SALT", "")
	hash := sha1.Sum([]byte(
		salt +
			param.Scope +
			param.Client_id +
			param.Redirect_uri +
			param.State +
			param.Nonce))
	return fmt.Sprintf("%x", hash)
}

func getIdToken(param Param, name, password string) (string, error) {
	iss := getEnv("IMPLICAUZANT_ISSUER", "Implicauzant")
	exp, _ := strconv.Atoi(getEnv("IMPLICAUZANT_EXPIRES_IN", "86400"))
	secret := []byte(getEnv("IMPLICAUZANT_SECRET_KEY", ""))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = getSub(name, password)
	claims["iss"] = iss
	claims["aud"] = param.Client_id
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Duration(exp) * time.Second).Unix()
	claims["nonce"] = param.Nonce
	if strings.Index(param.Scope, "profile") >= 0 {
		claims["name"] = name
	}
	id_token, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	url := &url.URL{}
	query := url.Query()
	query.Set("token_type", "Bearer")
	query.Set("id_token", id_token)
	query.Set("state", param.State)
	return query.Encode(), nil
}

func getSub(name, password string) string {
	salt := getEnv("IMPLICAUZANT_SALT", "")
	postfix := getEnv("IMPLICAUZANT_SUBJECT_POSTFIX", "@implicauzant")
	sub := sha1.Sum([]byte(name + password + salt))
	return fmt.Sprintf("%x%s", sub, postfix)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
