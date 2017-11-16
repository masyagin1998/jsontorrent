package main

import (
	"encoding/json"
	"github.com/mgutz/logxi/v1"
	"net"
)

// Peer struct contains peer information.
type Peer struct {
	// TCP
	TCPAddr     *net.TCPAddr
	isConnected bool
	conn        *net.TCPConn
	// JSON
	encoder    *json.Encoder
	decoder    *json.Decoder
	have       []bool
	waitingFor []bool
}

// InitPeer creates new peer.
func InitPeer(TCPAddr *net.TCPAddr, numberOfPieces int) *Peer {
	have := make([]bool, numberOfPieces)
	waitingFor := make([]bool, numberOfPieces)
	return &Peer{
		TCPAddr:     TCPAddr,
		isConnected: false,
		conn:        nil,
		encoder:     nil,
		decoder:     nil,
		have:        have,
		waitingFor:  waitingFor,
	}
}

// Join connects to peers.
func (server *Server) Join(conn *net.TCPConn, peer *Peer) {
	peer.conn = conn
	peer.encoder, peer.decoder = json.NewEncoder(conn), json.NewDecoder(conn)
	peer.isConnected = true
	go server.InPeer(peer)
}

// InPeer reads messages from other peers.
func (server *Server) InPeer(peer *Peer) {
	bitfield := make([]bool, len(server.fileInfo))
	for i := 0; i < len(server.fileInfo); i++ {
		bitfield[i] = server.fileInfo[i].have
	}
	err := peer.encoder.Encode(Message{Command: "bitfield", Bitfield: bitfield})
	if err != nil {
		server.KillPeer(peer)
		return
	}
	log.Info("successfully connetcted.", "to", peer.TCPAddr.String())
	for peer.isConnected {
		var msg Message
		err := peer.decoder.Decode(&msg)
		if err != nil {
			server.KillPeer(peer)
			return
		}
		server.messages <- ChanMessage{
			message: msg,
			peer:    peer,
		}
	}
}

// KillPeer kills peer in which there was an error.
func (server *Server) KillPeer(peer *Peer) {
	err := peer.conn.Close()
	if err != nil {
		log.Error("error, while closing connection.", "from", peer.TCPAddr.String(), "error", err)
	} else {
		peer.encoder = nil
		peer.decoder = nil
		peer.isConnected = false
		for i := 0; i < len(peer.have); i++ {
			peer.have[i] = false
			peer.waitingFor[i] = false
		}
		log.Info("successfully closed connection.", "from", peer.TCPAddr.String())
	}
}
