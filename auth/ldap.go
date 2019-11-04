package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tidwall/gjson"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

// LdapPlugin ..
type LdapPlugin struct {
	Credentials VaultCredentials
	VaultURL    string
}

// LdapCredentials ...
type LdapCredentials struct {
	Username string
	Password string
}

// Get ... just returns itself on calls to Get()
func (l LdapCredentials) Get() (interface{}, error) {
	return l, nil
}

// Auth ...
func (l *LdapPlugin) Auth(credential VaultCredentials) (string, error) {

	// verify the passed creds are LdapCredentials
	ldapCreds, ok := credential.(*LdapCredentials)
	if !ok {
		return "", errors.New("LdapPlugin.Auth() plugin can only acces VaultCredentials of type LdapCredentials")
	}

	// create our POST payload for the login
	jsonData, err := json.Marshal(map[string]string{"password": ldapCreds.Password})
	if err != nil {
		return "", errors.New("LdapPlugin.Auth() plugin error marshalling VaultCredentials: " + err.Error())
	}

	// convert to bytes
	jsonBytes := bytes.NewBuffer(jsonData)

	// new http client
	client := &http.Client{}
	url := (l.VaultURL + "/v1/auth/ldap/login/" + ldapCreds.Username)

	// sending POST for the VaultCredentials
	log.Info("LdapPlugin authenticating for: " + url)
	req, err := http.NewRequest("POST", url, jsonBytes)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("LdapPlugin.Auth() plugin error error logging into vault: " + err.Error())
	}
	defer resp.Body.Close()

	// get the clientToken
	var clientToken string

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		clientToken = gjson.Get(bodyString, "auth.client_token").String()
	} else {
		return "", errors.New("Vault authentication failed: " + http.StatusText(resp.StatusCode))
	}

	return clientToken, nil

}

// GetCredentials ...
func (l *LdapPlugin) GetCredentials(req *http.Request) (VaultCredentials, error) {

	// we require basic auth
	// and this is relayed to vault
	user, pass, ok := req.BasicAuth()

	if !ok {
		return nil, errors.New("Basic Auth is required")
	}

	l.Credentials = &LdapCredentials{Username: user, Password: pass}

	return l.Credentials, nil

}
