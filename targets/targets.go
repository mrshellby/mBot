package targets

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/un4gi/mBot/config"
	"github.com/un4gi/mBot/data"
	"github.com/un4gi/mBot/env"
	"github.com/un4gi/mBot/requests"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func CheckTargets(target string) {

	statuscode, body := requests.DoGetRequest(target)

	defer body.Close()
	if statuscode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		var newTargets data.TargetData
		err = json.Unmarshal(bodyBytes, &newTargets)
		bodyString := string(bodyBytes)

		if len(bodyString) > 3 { // Check to see if there is a JSON response
			for i := 0; i < len(newTargets); i++ { // Iterate over the number of unregistered targets
				if newTargets[i].Category.ID >= 3 { // Check to see if new target is a web/host target
				} else {
					if !newTargets[i].Registered {
						log.Printf(env.InfoColor, "Onboarding to "+newTargets[i].Codename+" - "+newTargets[i].Slug)
						OnboardTarget("https://platform.synack.com/api/targets/" + newTargets[i].Slug + "/signup")
					}
				}
			}
		}
	} else if statuscode == http.StatusUnauthorized {
		config.LoggedIn = false
		RenewSession()
	}
}

func OnboardTarget(target string) {
	var jsonStr = []byte(`{"ResearcherListing":{"terms":1}}`)
	requests.DoPostRequest(target, jsonStr)
}

func CheckForQR(target string) {
	statuscode, body := requests.DoGetRequest(target)

	defer body.Close()
	if statuscode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		var newTargets data.TargetData
		err = json.Unmarshal(bodyBytes, &newTargets)
		bodyString := string(bodyBytes)

		if len(bodyString) > 3 { // Check to see if there is a JSON response
			for i := 0; i < len(newTargets); i++ { // Iterate over the number of unregistered targets
				if newTargets[i].Category.ID >= 3 { // Check to see if new target is a web/host target
				} else {
					log.Printf(env.SuccessColor, "[!] Target "+newTargets[i].Codename+" is in QR!")
				}
			}
		}
	} else if statuscode == http.StatusUnauthorized {
		config.LoggedIn = false
		RenewSession()
	}
}

func RenewSession() {
	log.Printf(env.WarningColor, "Bearer token expired.")

	auth := []byte(`{"email":"` + config.Email + `","password":"` + config.Password + `"}`)
	if string(auth) == "" {
		os.Exit(1)
	}

	token, cookie := GetCSRFToken()
	pt := GenerateProgressToken(auth, token, cookie)

	log.Printf(env.DebugColor, "Generating TOTP Token...")
	key := GeneratePassCode()
	log.Printf(env.DebugColor, "Authenticating...")
	ValidateTFA(key, pt, token, cookie)
}

func GeneratePassCode() string {
	passcode, err := totp.GenerateCodeCustom(config.AuthySecret, time.Now(), totp.ValidateOpts{
		Period:    10,
		Skew:      1,
		Digits:    otp.Digits(7),
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		panic(err)
	}
	return passcode
}

func GetCSRFToken() (string, string) {
	var token string
	statuscode, body, headers := requests.DoLoginGetRequest("https://login.synack.com")
	defer body.Close()

	cookie := headers["Set-Cookie"][0]
	if statuscode != http.StatusEarlyHints {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		bodyString := string(bodyBytes)
		bodyLines := strings.Split(bodyString, "\n")
		if len(bodyLines) > 5 {
			for _, line := range bodyLines {
				if strings.Contains(string(line), `<meta name="csrf-token"`) {
					elems := strings.Split(string(line), `"`)
					token := elems[3]
					return token, cookie
				}
			}
		}
	}
	return token, cookie
}

func GenerateProgressToken(auth []byte, token string, cookie string) string {
	var progress_token string
	statuscode, body := requests.DoLoginPostRequest(requests.Urls[5], auth, token, cookie)
	defer body.Close()

	if statuscode != http.StatusForbidden {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		bodyString := string(bodyBytes)
		if len(bodyString) > 0 {
			elems := strings.Split(bodyString, `"`)
			if len(elems) >= 5 {
				progress_token = elems[5]
				return progress_token
			} else {
				log.Printf(env.ErrorColor, "Error requesting progress token.")
				progress_token = ""
				return progress_token
			}
		}
	}
	return progress_token
}

func ValidateTFA(key, pt, token, cookie string) {
	auth := []byte(`{"authy_token":"` + key + `","progress_token":"` + pt + `"}`)
	statuscode, body := requests.DoLoginPostRequest(requests.Urls[5], auth, token, cookie)
	defer body.Close()

	if statuscode != http.StatusForbidden {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		bodyString := string(bodyBytes)

		if len(bodyString) > 0 {
			elems := strings.Split(bodyString, `"`)
			if len(elems) >= 9 {
				grant_token := elems[9]
				GetSessionToken(grant_token)
			} else {
				log.Printf(env.ErrorColor, "Error requesting grant token.")
			}

		}
	}
}

func GetSessionToken(gt string) {
	statuscode, body := requests.DoGetRequest("https://platform.synack.com/?grant_token=" + gt)
	defer body.Close()

	if statuscode != http.StatusForbidden {
		log.Printf(env.DebugColor, "Grabbing access token...")
	}

	statuscode, body = requests.DoGrantTokenRequest("https://platform.synack.com/token?grant_token=" + gt)
	defer body.Close()

	if statuscode != http.StatusForbidden {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled {
			log.Printf(env.ErrorColor, err)
		}

		bodyString := string(bodyBytes)
		if len(bodyString) > 0 {
			elems := strings.Split(bodyString, `"`)
			if len(elems) >= 3 {
				access_token := elems[3]
				log.Printf(env.SuccessColor, "New access token: "+access_token)
				requests.Token = access_token
				config.LoggedIn = true
			} else {
				log.Printf(env.ErrorColor, "Error requesting access token.")
			}
		}
	}
}
