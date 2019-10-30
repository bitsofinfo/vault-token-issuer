package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/gorilla/mux"
)

var (
	vaultUrl = flag.String("vault-url", "", "Vault url")
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {

	flag.Parse()

	log.Info("Initializing vaulttoken issuer")

	r := mux.NewRouter()
	r.HandleFunc("/", TestHandler)
	http.Handle("/", r)

	pemCert, pemKey := generate(4096)

	cert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8443",
		TLSConfig:    cfg,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func TestHandler(w http.ResponseWriter, r *http.Request) {

	user, pass, ok := r.BasicAuth()

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	} else {

		body2, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		var createTokenPayload createOrphanPayload
		json.Unmarshal(body2, &createTokenPayload)

		x, err := json.Marshal(createTokenPayload)
		log.Info(string(x))

		if len(createTokenPayload.Policies) == 0 {
			http.Error(w, "Policies required", http.StatusBadRequest)
		} else {
			token, err := getToken(user, pass, &createTokenPayload)

			if err != nil {
				log.Info(err)
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "token: %s\n", token)
		}

	}

}

type createOrphanPayload struct {
	Renewable string   `json:"renewable"`
	Period    string   `json:"period"`
	Policies  []string `json:"policies"`
}

func getToken(username string, password string, createTokenPayload *createOrphanPayload) (string, error) {

	payload := map[string]string{"password": password}
	jsonData, err := json.Marshal(payload)
	jsonBytes := bytes.NewBuffer(jsonData)

	client := &http.Client{}
	url := (*vaultUrl + "/v1/auth/ldap/login/" + username)
	log.Info(url)
	req, err := http.NewRequest("POST", url, jsonBytes)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var clientToken gjson.Result

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		//log.Info(bodyString)

		clientToken = gjson.Get(bodyString, "auth.client_token")

		//fmt.Print(clientToken)
	}

	jsonData, err = json.Marshal(createTokenPayload)
	//log.Info(jsonData)
	jsonBytes = bytes.NewBuffer(jsonData)
	url = (*vaultUrl + "/v1/auth/token/create-orphan")
	log.Info(url)
	req, err = http.NewRequest("POST", url, jsonBytes)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", clientToken.String())
	resp, err = client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		//log.Info(bodyString)

		clientToken = gjson.Get(bodyString, "auth.client_token")

		//fmt.Print(clientToken)

		return clientToken.String(), nil
	}

	return "", errors.New("failed to get token")
}
