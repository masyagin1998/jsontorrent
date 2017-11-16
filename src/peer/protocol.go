package main

// Message struct is used to peer's communication.
type Message struct {
	Command  string `json:"command"`
	Have     int    `json:"have"`
	Bitfield []bool `json:"bitfield"`
	Request  int    `json:"request"`
	Piece    []byte `json:"piece"`
	Index    int    `json:"index"`
}
