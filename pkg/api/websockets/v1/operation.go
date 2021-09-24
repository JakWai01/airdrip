package api

type Application struct {
	Message
	Community string `json:"community"`
	Mac       string `json:"mac"`
}

type Acceptance struct {
	Message
}

type Rejection struct {
	Message
}

type Ready struct {
	Message
	Mac string `json:"mac"`
}

type Introduction struct {
	Message
	Mac string `json:"mac"`
}

type Offer struct {
	Message
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Answer struct {
	Message
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Candidate struct {
	Message
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Exited struct {
	Message
}

type Resignation struct {
	Message
	Mac string `json:"mac"`
}

func NewApplication(community string, mac string) *Application {
	return &Application{Message: Message{OpcodeApplication}, Community: community, Mac: mac}
}

func NewAcceptance() *Acceptance {
	return &Acceptance{Message: Message{OpcodeAcceptance}}
}

func NewRejection() *Rejection {
	return &Rejection{Message: Message{OpcodeRejection}}
}

func NewReady(mac string) *Ready {
	return &Ready{Message: Message{OpcodeReady}, Mac: mac}
}

func NewIntroduction(mac string) *Introduction {
	return &Introduction{Message: Message{OpcodeIntroduction}, Mac: mac}
}

func NewOffer(mac string, payload string) *Offer {
	return &Offer{Message: Message{OpcodeOffer}, Mac: mac, Payload: payload}
}

func NewAnswer(mac string, payload string) *Answer {
	return &Answer{Message: Message{OpcodeAnswer}, Mac: mac, Payload: payload}
}

func NewCandidate(mac string, payload string) *Candidate {
	return &Candidate{Message: Message{OpcodeCandidate}, Mac: mac, Payload: payload}
}

func NewExited() *Exited {
	return &Exited{Message: Message{OpcodeExited}}
}

func NewResignation(mac string) *Resignation {
	return &Resignation{Message: Message{OpcodeResignation}, Mac: mac}
}