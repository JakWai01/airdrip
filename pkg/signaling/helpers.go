package signaling

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

func (s *SignalingServer) getCommunity(mac string) (string, error) {
	for key, element := range s.communities {
		for i := 0; i < len(element); i++ {
			if element[i] == mac {
				return key, nil
			}
		}
	}

	return "", errors.New("This mac is not part of any community so far!")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func deleteElement(s []string, str string) []string {
	var elementIndex int
	for index, element := range s {
		if element == str {
			elementIndex = index
		}
	}
	return append(s[:elementIndex], s[elementIndex+1:]...)
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *SignalingServer) getSenderMac(receiverMac string, community string) string {
	if len(s.communities[community]) == 2 {
		if receiverMac == s.communities[community][1] {
			// The second one is sender
			return s.communities[community][0]
		} else {
			// First one
			return s.communities[community][1]
		}
	} else {
		return s.communities[community][1]
	}
}
