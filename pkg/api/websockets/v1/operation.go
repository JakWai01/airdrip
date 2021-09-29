package api

type Application struct {
	Message
	Community string `json:"community"`
}

type Acceptance struct {
	Message
}

type Rejection struct {
	Message
}

type Ready struct {
	Message
}

type Introduction struct {
	Message
}

type Offer struct {
	Message
	Payload []byte `json:"payload"`
}

type Answer struct {
	Message
	Payload []byte `json:"payload"`
}

type Candidate struct {
	Message
	Payload []byte `json:"payload"`
}

type Exited struct {
	Message
}

type Resignation struct {
	Message
}

func NewApplication(community string) *Application {
	return &Application{Message: Message{OpcodeApplication}, Community: community}
}

func NewAcceptance() *Acceptance {
	return &Acceptance{Message: Message{OpcodeAcceptance}}
}

func NewRejection() *Rejection {
	return &Rejection{Message: Message{OpcodeRejection}}
}

func NewReady() *Ready {
	return &Ready{Message: Message{OpcodeReady}}
}

func NewIntroduction() *Introduction {
	return &Introduction{Message: Message{OpcodeIntroduction}}
}

func NewOffer(payload []byte) *Offer {
	return &Offer{Message: Message{OpcodeOffer}, Payload: payload}
}

func NewAnswer(payload []byte) *Answer {
	return &Answer{Message: Message{OpcodeAnswer}, Payload: payload}
}

func NewCandidate(payload []byte) *Candidate {
	return &Candidate{Message: Message{OpcodeCandidate}, Payload: payload}
}

func NewExited() *Exited {
	return &Exited{Message: Message{OpcodeExited}}
}

func NewResignation() *Resignation {
	return &Resignation{Message: Message{OpcodeResignation}}
}
