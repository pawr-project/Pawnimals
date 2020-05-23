package controller

import (
	"encoding/json"
	"math"
	"time"

	"github.com/appditto/natricon/server/db"
	"github.com/appditto/natricon/server/image"
	"github.com/appditto/natricon/server/model"
	"github.com/appditto/natricon/server/net"
	"github.com/appditto/natricon/server/utils"
	"github.com/bsm/redislock"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Account donations are checked to
const donationAccount = "nano_1natrium1o3z5519ifou7xii8crpxpk8y65qmkih8e8bpsjri651oza8imdd"

// Donations at or above this threshold will award "vip" status for 30 days
const donationThresholdNano = 2.0

// Maximum donation days allowed at a time (cannot buy bigger chunks than this at a time)
const maxVIPDays = 120

type NanoController struct {
	RPCClient *net.RPCClient
}

// Handle callback for donation listener
func (nc NanoController) Callback(c *gin.Context) {
	var callbackData model.Callback
	err := c.BindJSON(&callbackData)
	if err != nil {
		glog.Errorf("Error processing callback")
		return
	}
	var blockData model.Block
	err = json.Unmarshal([]byte(callbackData.Block), &blockData)
	if err != nil {
		glog.Errorf("Error parsing callback block json")
		return
	}
	// Check if send to doantion account
	if blockData.LinkAsAccount == donationAccount && blockData.LinkAsAccount != blockData.Account {
		durationDays := nc.calcDonorDurationDays(callbackData.Amount)
		if durationDays > 0 {
			glog.Infof("Giving donor status to %s for %d days", blockData.Account, durationDays)
			db.GetDB().UpdateDonorStatus(callbackData.Hash, blockData.Account, durationDays, maxVIPDays)
		}
	}
}

// Cron job for checking missed callbacks
func (nc NanoController) CheckMissedCallbacks() {
	if nc.RPCClient == nil {
		return
	}

	// Try to obtain lock.
	lock, err := db.GetDB().Locker.Obtain("natricon:history_lock", 100*time.Second, nil)
	if err == redislock.ErrNotObtained {
		return
	} else if err != nil {
		glog.Error(err)
		return
	}
	defer lock.Release()
	// Check history
	historyResponse, err := nc.RPCClient.MakeAccountHistoryRequest(
		donationAccount,
		10,
	)
	if err != nil {
		glog.Errorf("Error occured checking donation account history %s", err)
		return
	}
	for i := 0; i < len(historyResponse.History); i++ {
		if historyResponse.History[i].Type == "receive" && historyResponse.History[i].Account != donationAccount {
			durationDays := nc.calcDonorDurationDays(historyResponse.History[i].Amount)
			if durationDays > 0 {
				glog.Infof("Checking donor status to %s for %d days", historyResponse.History[i].Account, durationDays)
				db.GetDB().UpdateDonorStatus(historyResponse.History[i].Hash, historyResponse.History[i].Account, durationDays, maxVIPDays)
			}
		}
	}
}

// calcDonorDurationDays - calculate how long badge will persist with given donation amount
func (nc NanoController) calcDonorDurationDays(amountRaw string) uint {
	amountNano, _ := utils.RawToNano(amountRaw)
	chunks := uint(amountNano / donationThresholdNano)
	return uint(math.Min(float64(chunks*30), maxVIPDays))
}

// Cron job for updating principal rep weight requirement
func (nc NanoController) UpdatePrincipalWeight() {
	if nc.RPCClient == nil {
		return
	}
	// Check history
	quorumResponse, err := nc.RPCClient.MakeConfirmationQuorumRequest()
	if err != nil {
		glog.Errorf("Error occured checking confirmation quorum %s", err)
		return
	}
	onlineWeightMinimum, err := utils.RawToNano(quorumResponse.OnlineWeightTotal)
	if err != nil {
		glog.Errorf("Error occured converting weight to nano %s", err)
		return
	}
	// 1% of online weight means principal rep
	principalRepMinimum := onlineWeightMinimum * 0.01
	glog.Infof("Setting principal rep requirement to %f", principalRepMinimum)
	db.GetDB().SetPrincipalRepRequirement(principalRepMinimum)
}

// Cron job for updating principal reps
func (nc NanoController) UpdatePrincipalReps() {
	if nc.RPCClient == nil {
		return
	}
	// Get weight requirement
	repWeightRequirement := db.GetDB().GetPrincipalRepRequirement()
	// Get reps
	repsResponse, err := nc.RPCClient.MakeRepresentativesRequest()
	if err != nil {
		glog.Errorf("Error occured checking confirmation quorum %s", err)
		return
	}
	principalReps := []string{}
	for rep, weight := range repsResponse.Representatives {
		weightNano, err := utils.RawToNano(weight)
		if err != nil {
			continue
		}
		if weightNano >= repWeightRequirement {
			principalReps = append(principalReps, utils.AddressToPub(rep))
		}
	}
	// Update cache
	db.GetDB().SetPrincipalReps(principalReps)
	// Update badge service
	image.GetBadgeSvc().PrincipalReps = principalReps
}
