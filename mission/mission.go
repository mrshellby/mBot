package mission

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/un4gi/mBot/config"
	"github.com/un4gi/mBot/data"
	"github.com/un4gi/mBot/discord"
	"github.com/un4gi/mBot/env"
	"github.com/un4gi/mBot/requests"
	"github.com/un4gi/mBot/targets"
	"github.com/mrshellby/mBot-dev/slack"
)

var blacklist []string

func CheckBlacklist(str string) bool {
	for _, v := range blacklist {
		if v == str {
			return true
		}
	}
	return false
}

func CheckClaimed(target string, header []string, missionurl string) {
	statuscode, body := requests.DoGetRequest(target)
	defer body.Close()

	if statuscode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled && err != io.EOF {
			log.Println(err)
			return
		}
		var claimedMissions data.MissionDataV2
		err = json.Unmarshal(bodyBytes, &claimedMissions)
		if err != nil && err != context.Canceled && err != io.EOF {
			fmt.Println("Error unmarshalling data:", err)
		}
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		lenResp := len(claimedMissions)
		if len(bodyString) > 3 { // Check to see if there is a JSON response
			fmt.Println("Starting loop to attempt to grab missions...")
			for i := 0; i < lenResp; i++ {
				// v1 - missionurl + "/organizations/"+newMissions[i].Organization.ID+"/listings/"+newMissions[i].Listing.ID+"/campaigns/"+newMissions[i].Campaign.ID+"/tasks/"+newMissions[i].ID+"/transitions
				// v2 - missionurl + claimedMissions[i].ID + "/evidences"
				missionEditURL := missionurl + claimedMissions[i].ID + "/evidences"
				EditMission(missionEditURL, header)
			}
		}
	} else if statuscode == http.StatusUnauthorized {
		config.LoggedIn = false
		targets.RenewSession()
	}
}

func CheckMissions(target string) {
	statuscode, body := requests.DoGetRequest(target)
	defer body.Close()

	if statuscode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled && err != io.EOF {
			log.Println(err)
			return
		}

		//var newMissions data.MissionData
		var newMissions data.MissionDataV2
		err = json.Unmarshal(bodyBytes, &newMissions)
		bodyString := string(bodyBytes)
		lenResp := len(newMissions)
		if len(bodyString) > 3 { // Check to see if there is a JSON response
			var wg sync.WaitGroup
			wg.Add(lenResp)

			fmt.Println("Starting loop to attempt to grab missions...")
			for i := 0; i < lenResp; i++ {
				go func(i int) {
					defer wg.Done()
					// API v1
					// log.Println("Saw", newMissions[i].Payout.Amount, newMissions[i].Listing.Title, "mission -", newMissions[i].Title)
					// GrabMission("https://platform.synack.com/api/tasks/v1/organizations/"+newMissions[i].Organization.ID+"/listings/"+newMissions[i].Listing.ID+"/campaigns/"+newMissions[i].Campaign.ID+"/tasks/"+newMissions[i].ID+"/transitions", newMissions[i].Payout.Amount, newMissions[i].Listing.Title, newMissions[i].Title, 10)
					// log.Println("Attempted to grab", newMissions[i].Payout.Amount, newMissions[i].Listing.Title, "mission -", newMissions[i].Title)

					// API v2
					log.Println("Saw", newMissions[i].Payout.Amount, newMissions[i].ListingCodename, "mission -", newMissions[i].Title)

					if strings.Contains(newMissions[i].ListingCodename, "DAISY") {
						blacklist = append(blacklist, newMissions[i].ListingCodename)
						log.Printf(env.WarningColor, "Added "+newMissions[i].ListingCodename+" to the blacklist.")
					}
					// API v1 with API v2 references
					if !CheckBlacklist(newMissions[i].ListingCodename) {
						GrabMission("https://platform.synack.com/api/tasks/v1/organizations/"+newMissions[i].OrganizationUid+"/listings/"+newMissions[i].ListingUid+"/campaigns/"+newMissions[i].CampaignUid+"/tasks/"+newMissions[i].ID+"/transitions", newMissions[i].Payout.Amount, newMissions[i].ListingCodename, newMissions[i].Title, 10)

						// API v2
						//GrabMission("https://platform.synack.com/api/tasks/v2/organizations/"+newMissions[i].OrganizationUid+"/listings/"+newMissions[i].ListingUid+"/campaigns/"+newMissions[i].CampaignUid+"/tasks/"+newMissions[i].ID+"/transitions", newMissions[i].Payout.Amount, newMissions[i].ListingCodename, newMissions[i].Title, 10)
						log.Println("Attempted to grab", newMissions[i].Payout.Amount, newMissions[i].ListingCodename, "mission -", newMissions[i].Title)
					} else {
						log.Println("Maximum number of missions already claimed for", newMissions[i].ListingCodename+". Skipping mission.")
					}
				}(i)
			}
			wg.Wait()
			fmt.Println("Finished attempt to grab new missions...")
		}
	} else if statuscode == http.StatusUnauthorized {
		config.LoggedIn = false
		targets.RenewSession()
	}
}

func EditMission(url string, header []string) {
	//payload := data.CrossOriginResourceSharing("", false)

	//client := &http.Client{}
	//req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	//requests.SetHeaders(req)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//resp, err := client.Do(req)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(string(body))
}

func GrabMission(target string, pay int, project string, title string, n int) {
	var jsonStr = []byte(`{"type": "CLAIM"}`)
	statuscode, body := requests.DoPostRequest(target, jsonStr)
	defer body.Close()
	if statuscode == http.StatusCreated {
		bodyBytes, err := ioutil.ReadAll(body)
		if err != nil && err != context.Canceled && err != io.EOF {
			log.Println(err)
			return
		}

		bodyString := string(bodyBytes)
		if len(bodyString) > 0 { // Check to see if there is a JSON response
			log.Printf(env.SuccessColor, "Grabbed $ "+fmt.Sprint(pay)+" "+project+" mission - "+title)
			discord.ConnectDiscord("Grabbed $ " + fmt.Sprint(pay) + " " + project + " mission - " + title)
			slack.Hook("Grabbed $ " + fmt.Sprint(pay) + " " + project + " mission - " + title)
		} else {
			log.Printf(env.WarningColor, "Error grabbing $"+fmt.Sprint(pay)+" "+project+" mission - "+title)
		}
	} else if statuscode == http.StatusUnauthorized {
		log.Printf(env.WarningColor, "Bearer token expired.")
		os.Exit(0)
	} else if statuscode == http.StatusInternalServerError {
		if n > 0 {
			log.Printf(env.WarningColor, "500 Error. Retrying...")
			n = n - 1
			GrabMission(target, pay, project, title, n)
		} else {
			log.Printf(env.ErrorColor, "500 Error. 10 attempts have been made. Giving up!")
		}
	} else if statuscode == http.StatusForbidden {
		log.Printf(env.ErrorColor, "Too slow! Someone grabbed it.")
	} else if statuscode == http.StatusPreconditionFailed {
		log.Printf(env.ErrorColor, "You already have the maximum number of missions for this target.")
		blacklist = append(blacklist, project)
	} else {
		log.Printf(env.WarningColor, "Status Code: "+fmt.Sprint(statuscode))
	}
	body.Close()
}
