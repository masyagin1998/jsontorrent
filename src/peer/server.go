package main

import (
	"github.com/mgutz/logxi/v1"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// FilePiece struct contains information about one piece of file.
type FilePiece struct {
	data string
	have bool
}

// Server struct contains all peer data.
type Server struct {
	// Peers.
	myPeer     *Peer
	otherPeers []*Peer
	// Torrent.
	filename string
	fileInfo []FilePiece
	haveFile bool
	// Messages.
	messages chan ChanMessage
}

// ChanMessage struct is used to select work.
type ChanMessage struct {
	message Message
	peer    *Peer
}

// NewServer creates new server.
func NewServer(myIP *net.TCPAddr, otherIPs []*net.TCPAddr,
	seedFilename string,
	filename string, length, pieceLength, numberOfPieces int) *Server {
	// Peers.
	myPeer := InitPeer(myIP, numberOfPieces)
	otherPeers := make([]*Peer, len(otherIPs))
	for i := 0; i < len(otherIPs); i++ {
		otherPeers[i] = InitPeer(otherIPs[i], numberOfPieces)
	}
	// Torrent.
	fileInfo := make([]FilePiece, numberOfPieces)
	haveFile := false
	if seedFilename != "" {
		haveFile = true
		file, err := os.Open(seedFilename)
		if err == nil {
			buffer := make([]byte, pieceLength)
			for i := 0; i < pieceLength; i++ {
				n, err := file.Read(buffer)
				if (err != nil) && (err != io.EOF) {
					break
				}
				err = nil
				if n > 0 {
					fileInfo[i].data = string(buffer[:n])
					fileInfo[i].have = true
				}
			}
		}
		err = file.Close()
		log.Info("Succesfully read " + seedFilename)
	}
	// Messages.
	messages := make(chan (ChanMessage))
	return &Server{
		myPeer:     myPeer,
		otherPeers: otherPeers,
		filename:   filename,
		fileInfo:   fileInfo,
		haveFile:   haveFile,
		messages:   messages,
	}
}

// RunServer runs server.
func (server *Server) RunServer() {
	go server.ListenLoop()
	server.SearchPeers()
	server.SelectLoop()
}

// ListenLoop listens for new connections.
func (server *Server) ListenLoop() {
	listener, err := net.ListenTCP("tcp", server.myPeer.TCPAddr)
	if err != nil {
		log.Error("error, while starting listener.", "error", err)
		return
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Error("error, while accepting connection.", "error", err)
		} else {
			IPLength := make([]byte, 1)
			_, err := conn.Read(IPLength)
			if err != nil {
				log.Error("error, while getting IP adress.", "error", err)
			} else {
				IPString := make([]byte, IPLength[0])
				_, err := conn.Read(IPString)
				if err != nil {
					log.Error("error, while getting IP adress.", "error", err)
				} else {
					IPString := string(IPString)
					flag := false
					for _, peer := range server.otherPeers {
						if peer.TCPAddr.String() == IPString {
							server.Join(conn, peer)
							flag = true
							break
						}
					}
					if !flag {
						log.Info("unknown peer tried to connect.")
						err = conn.Close()
						if err != nil {
							log.Error("error, while closing connection from unknown peer", "error", err)
						}
					}
				}
			}
		}
	}
}

// SearchPeers trys to connect to known peers.
func (server *Server) SearchPeers() {
	for _, peer := range server.otherPeers {
		if !peer.isConnected {
			conn, err := net.DialTCP("tcp", nil, peer.TCPAddr)
			if err == nil {
				server.Join(conn, peer)
				_, err = conn.Write([]byte{byte(len(peer.TCPAddr.String()))})
				if err != nil {
					log.Error("error, while sending IP adress", "error", err)
				} else {
					_, err = conn.Write([]byte(server.myPeer.TCPAddr.String()))
					if err != nil {
						log.Error("error, while sending IP adress", "error", err)
					}
				}
			}
		}
	}
}

// SelectLoop is a "Heart" of server.
func (server *Server) SelectLoop() {
	// Infinite Loop.
	for {
		select {
		case newMessage := <-server.messages:
			switch newMessage.message.Command {
			case "bitfield":
				newMessage.peer.have = newMessage.message.Bitfield
				log.Info("new message.",
					"from", newMessage.peer.TCPAddr.String(),
					"type", newMessage.message.Command,
					"bitfield", newMessage.peer.have)
			case "have":
				newMessage.peer.have[newMessage.message.Have] = true
				log.Info("new message.",
					"from", newMessage.peer.TCPAddr.String(),
					"type", newMessage.message.Command,
					"index of piece", newMessage.message.Have)
			case "request":
				log.Info("new message.",
					"from", newMessage.peer.TCPAddr.String(),
					"type", newMessage.message.Command,
					"index of piece", newMessage.message.Request)
				err := newMessage.peer.encoder.Encode(Message{Command: "piece", Piece: []byte(server.fileInfo[newMessage.message.Request].data), Index: newMessage.message.Request})
				if err != nil {
					server.KillPeer(newMessage.peer)
				}
			case "piece":
				log.Info("new message.",
					"from", newMessage.peer.TCPAddr.String(),
					"type", newMessage.message.Command,
					"index of piece", newMessage.message.Index)
				if !server.haveFile && !server.fileInfo[newMessage.message.Index].have {
					server.fileInfo[newMessage.message.Index].data = string(newMessage.message.Piece[:len(newMessage.message.Piece)])
					server.fileInfo[newMessage.message.Index].have = true
				}
				for _, peer := range server.otherPeers {
					if peer.isConnected {
						err := peer.encoder.Encode(Message{Command: "have", Have: newMessage.message.Index})
						if err != nil {
							server.KillPeer(peer)
						}
					}
				}
			}
		}
		if !server.haveFile {
			flag := true
			for _, pieceInfo := range server.fileInfo {
				if !pieceInfo.have {
					flag = false
					break
				}
			}
			if flag {
				// Touch file.
				file, err := os.Create(server.filename)
				if err != nil {
					log.Error("error, while creating file", "error", err)
					return
				}
				for i := 0; i < len(server.fileInfo); i++ {
					_, err = io.Copy(file, strings.NewReader(server.fileInfo[i].data))
					if err != nil {
						log.Error("error, while creating file", "error", err)
						return
					}
				}
				err = file.Close()
				if err != nil {
					log.Error("error, while creating file", "error", err)
					return
				}
				log.Info("hooray, torrent was succesfully downloaded!")
				server.haveFile = true
				continue
			}
			server.Request()
		}
	}
}

// Request makes request to another server.
func (server *Server) Request() {
	// Heuristics "First Rare".

	// Get rarest piece indexes.
	minimum := len(server.otherPeers) + 1
	index := -1
	indexes := make([]int, 0)
	for i := 0; i < len(server.fileInfo); i++ {
		localMinimum := 0
		for _, peer := range server.otherPeers {
			if (peer.isConnected) && (peer.have[i]) && (!peer.waitingFor[i]) && (!server.fileInfo[i].have) {
				localMinimum++
			} else if peer.waitingFor[i] {
				localMinimum = -1
				break
			}
		}
		if (localMinimum > 0) && (localMinimum < minimum) {
			minimum = localMinimum
			index = i
			indexes = nil
			indexes = make([]int, 0)
			indexes = append(indexes, index)
		} else if localMinimum == minimum {
			indexes = append(indexes, i)
		}
	}
	if index == -1 {
		return
	}
	// Chose rarest piece index by random.
	rand.Seed(time.Now().UTC().UnixNano())
	index = rand.Intn(len(indexes))
	index = indexes[index]

	// Get all peers which have rarest piece.
	peers := make([]*Peer, 0)
	for _, peer := range server.otherPeers {
		if (peer.isConnected) && (peer.have[index]) && (!peer.waitingFor[index]) {
			peers = append(peers, peer)
		} else if peer.waitingFor[index] {
			return
		}
	}
	// Chose peer for request by random.
	peerIndex := rand.Intn(len(peers))
	peers[peerIndex].waitingFor[index] = true
	err := peers[peerIndex].encoder.Encode(Message{Command: "request", Request: index})
	if err != nil {
		server.KillPeer(peers[peerIndex])
	} else {
		log.Info("request was sent successfully", "to", peers[peerIndex].TCPAddr.String(), "index of piece", index)
	}
}
