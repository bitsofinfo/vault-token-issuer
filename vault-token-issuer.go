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

	"github.com/asaskevich/govalidator"
	"github.com/bitsofinfo/vault-token-issuer/auth"
	"github.com/bitsofinfo/vault-token-issuer/util"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/gorilla/mux"
)

var (
	vaultUrl           string                  // vault URL (flag opt)
	vaultAuthenticator string                  // string vault auth plugin name (flag opt)
	authenticator      auth.VaultAuthenticator // the vault auth plugin
)

// the JSON payload we both consume
// from callers and relay to Vault
type createTokenPayload struct {
	Renewable bool     `valid:"required" json:"renewable"`
	Period    string   `valid:"alphanum,required" json:"period"`
	Policies  []string `valid:"ascii,required" json:"policies"`
}

// Struct for JSON we retun to caller
type createTokenResponse struct {
	Code  string `json:"code"`
	Token string `json:"token"`
	Msg   string `json:"msg"`
}

func init() {

	// cmd line args
	flag.StringVar(&vaultUrl, "vault-url", "", "Vault url where token API calls will be made.")
	flag.StringVar(&vaultAuthenticator, "vault-authenticator", "", "The vault authenticator plugin to use: valid options: 'ldap'")

	// logging options
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {

	flag.Parse()

	// load our plugin for the authentication backend
	if vaultAuthenticator == "ldap" {
		authenticator = &auth.LdapPlugin{VaultUrl: vaultUrl}
	} else {
		log.Fatal("Invalid --vault-authenticator specified. We only support 'ldap'")
	}

	// setup our REST routes
	router := mux.NewRouter()
	router.Path("/token/create-orphan").
		Methods("POST").
		Schemes("https").
		Headers("Content-Type", "application/json").
		HandlerFunc(CreateOrphanTokenHandler)

	// fire up the http server
	srv := &http.Server{
		Handler: router,
		Addr:    ":8443",

		// https://golang.org/doc/go1.6#http2 needed to disable 128b ciphers below which are required by h2
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),

		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{getTLSCertificate()},
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			}},
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func getTLSCertificate() tls.Certificate {
	// generate a self signed cert & key
	pemCert, pemKey := util.Generate(4096)

	// get cert from keypair
	cert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		log.Fatal("Unexpected error generating self signed cert", err)
	}

	return cert
}

func writeHttpResponse(resWriter http.ResponseWriter, code string, token string, msg string, httpStatus int) {
	resWriter.Header().Set("Content-Type", "application/json")
	resWriter.WriteHeader(httpStatus)
	json.NewEncoder(resWriter).Encode(&createTokenResponse{Code: code, Token: token, Msg: msg})
}

func CreateOrphanTokenHandler(resWriter http.ResponseWriter, req *http.Request) {

	// first lets get the credentials off the request
	vaultCredentials, err := authenticator.GetCredentials(req)
	if err != nil {
		writeHttpResponse(resWriter, "error", "", "Bad Request: auth required", http.StatusBadRequest)
		return
	}

	// lets get our createTokenPayload struct
	payload, err := extractCreateTokenPayload(&resWriter, req)
	if err != nil {
		log.Error("Invalid payload: " + err.Error())
		writeHttpResponse(resWriter, "error", "", err.Error(), http.StatusUnauthorized)
		return
	}

	// must have at least one policy
	if len(payload.Policies) == 0 {
		writeHttpResponse(resWriter, "error", "", "one or more vault 'policies' are required", http.StatusBadRequest)
		return

		// otherwise lets proceed
	} else {

		// auth the actual user against value and get
		// the client access/auth token which we can then
		// use to create the actual orphan token
		userToken, err := authenticator.Auth(vaultCredentials)
		if err != nil {
			log.Error("Failed to authenticated againsg vault w/ VaultCredentials: " + err.Error())
			writeHttpResponse(resWriter, "error", "", err.Error(), http.StatusUnauthorized)
			return
		}

		token, err := createOrphanToken(userToken, payload)
		if err != nil {
			log.Error("Failed to create orphan token: " + err.Error())
			writeHttpResponse(resWriter, "error", "", err.Error(), http.StatusInternalServerError)
			return
		}

		writeHttpResponse(resWriter, "ok", token,
			fmt.Sprintf("renewable:%v period:%v policies:%v",
				payload.Renewable,
				payload.Period,
				payload.Policies), http.StatusOK)
	}

}

// Extracts the createTokenPayload JSON payload from the Request
func extractCreateTokenPayload(resWriter *http.ResponseWriter, req *http.Request) (*createTokenPayload, error) {

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, errors.New("failed to read request body json")
	}
	var payload createTokenPayload
	json.Unmarshal(body, &payload)

	x, err := json.Marshal(&payload)
	log.Info(string(x))

	// validate
	_, err = govalidator.ValidateStruct(&payload)
	if err != nil {
		return nil, errors.New("JSON payload failed validation: " + err.Error())
	}

	return &payload, nil

}

func createOrphanToken(userToken string, payload *createTokenPayload) (string, error) {

	jsonData, err := json.Marshal(payload)
	jsonBytes := bytes.NewBuffer(jsonData)
	url := (vaultUrl + "/v1/auth/token/create-orphan")
	log.Info(url)
	req, err := http.NewRequest("POST", url, jsonBytes)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", userToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var clientToken string

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		//log.Info(bodyString)

		clientToken = gjson.Get(bodyString, "auth.client_token").String()

		//fmt.Print(clientToken)

		return clientToken, nil
	}

	return "", errors.New("failed to get token")
}
