package v2rayApi

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/crossfw/Air-Universe/pkg/structures"
	regexp "github.com/dlclark/regexp2"
	"io"
	"os"
	"strconv"

	"strings"
)

/*
regex:
UserID:		(?<=email:)\d*(?=-)
UserTag:	(?<=email:\d*-)\w*
UserIP:		((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}(?=.*accepted)
*/

type singleUserRecord struct {
	userId  uint32
	userTag string
	userIP  string
}

func (useRec *singleUserRecord) addIP(userIPs *[]structures.UserIP) error {
	addUserFlag := true
	for id := 0; id < len(*userIPs); id++ {
		if (*userIPs)[id].Id == useRec.userId && (*userIPs)[id].InTag == useRec.userTag {
			addUserFlag = false
			addIPFlag := true
			for _, userIP := range (*userIPs)[id].AliveIP {
				if userIP == useRec.userIP {
					addIPFlag = false
					break
				}
			}
			if addIPFlag == true {
				(*userIPs)[id].AliveIP = append((*userIPs)[id].AliveIP, useRec.userIP)
			}
		}
	}
	if addUserFlag == true {
		singleUserIP := structures.UserIP{
			Id:      useRec.userId,
			InTag:   useRec.userTag,
			AliveIP: []string{useRec.userIP},
		}
		*userIPs = append(*userIPs, singleUserIP)
	}
	return nil
}

func captureDetail(line string) (useRec singleUserRecord, err error) {

	reUserID, _ := regexp.Compile("(?<=email:)\\d*(?=-)", 1)
	reUserTag, _ := regexp.Compile("(?<=email:\\d*-)\\w*", 1)
	reUserIP, _ := regexp.Compile("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}(?=.*accepted)", 1)

	mUserID, _ := reUserID.FindStringMatch(line)
	mUserTag, _ := reUserTag.FindStringMatch(line)
	mUserIP, _ := reUserIP.FindStringMatch(line)

	if mUserIP == nil || mUserTag == nil || mUserID == nil {
		err = errors.New("can't match")
		return
	} else {
		userId, _ := strconv.Atoi(mUserID.String())
		useRec.userId = uint32(userId)
		useRec.userTag = mUserTag.String()
		useRec.userIP = mUserIP.String()
		return
	}
}

func ReadV2Log(baseCfg *structures.BaseConfig) (userIPs *[]structures.UserIP, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("model FindUserDiffer cause error - %s", r))
		}
	}()

	userIPs = new([]structures.UserIP)

	v2Log, err := os.OpenFile(baseCfg.Proxy.LogPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer v2Log.Close()

	buf := bufio.NewReader(v2Log)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		line = strings.TrimSpace(line)
		//fmt.Println(line)
		// Process code

		singUser, userErr := captureDetail(line)
		if userErr != nil {
			continue
		} else {
			_ = singUser.addIP(userIPs)
		}
	}

	// clear log
	clearLog, err := os.OpenFile(baseCfg.Proxy.LogPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Clear log error!", err)
		return
	}
	defer clearLog.Close()

	return
}
