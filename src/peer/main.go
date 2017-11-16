package main

import (
	"github.com/mgutz/logxi/v1"
)

func main() {
	// Read config file.
	configData, err := ReadConfig()
	if err != nil {
		log.Error("error, while reading config file.", "error", err)
		return
	}
	if configData.seedFilename != "" {
		log.Info("config file was succesfully read.",
			"my IP", configData.myIPString,
			"other IPs", configData.otherIPsString,
			"torrent filename", configData.torrentFilename,
			"seed filename", configData.seedFilename)
	} else {
		log.Info("config file was succesfully read.",
			"my IP", configData.myIPString,
			"other IPs", configData.otherIPsString,
			"torrent filename", configData.torrentFilename)
	}
	// Read torrent file.
	torrentData, err := ReadTorrent(configData.torrentFilename)
	if err != nil {
		log.Error("error, while reading .torrent file.", "error", err)
		return
	}
	log.Info(".torrent file was succesfully read.",
		"name", torrentData.filename,
		"length", torrentData.length,
		"piece length", torrentData.pieceLength,
		"number of pieces", torrentData.numberOfPieces)
	// Run server.
	server := NewServer(configData.myIP, configData.otherIPs,
		configData.seedFilename,
		torrentData.filename, torrentData.length, torrentData.pieceLength, torrentData.numberOfPieces)
	server.RunServer()
}
