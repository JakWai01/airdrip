package signaling

import "errors"

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
