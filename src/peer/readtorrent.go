package main

import (
	"io"
	"os"
	"strconv"
	"strings"
)

// TorrentData is used to create peer.
type TorrentData struct {
	filename       string
	length         int
	pieceLength    int
	numberOfPieces int
}

// ReadTorrent gets metainformation from .torrent file.
func ReadTorrent(torrentFilename string) (data TorrentData, err error) {
	// Open config file.
	file, err := os.Open(torrentFilename)
	if err != nil {
		return data, err
	}
	// Read torrent file.
	buffer := make([]byte, 32768)
	bufferSize, err := file.Read(buffer)
	if (err != nil) && (err != io.EOF) {
		return data, err
	}
	file.Close()
	metaInfo := string(buffer[:bufferSize])
	i := 0
	j := 0
	// Find seed filename.
	i = strings.Index(metaInfo, "4:name")
	i += 6
	j = i
	for ; isCyfer(metaInfo[i]); i++ {
	}
	nameLength, err := strconv.Atoi(metaInfo[j:i])
	if err != nil {
		return data, err
	}
	data.filename = metaInfo[(i + 1):(i + 1 + nameLength)]
	// Find seed length.
	i = strings.Index(metaInfo, "6:lengthi")
	i += 9
	j = i
	for ; metaInfo[i] != 'e'; i++ {
	}
	data.length, err = strconv.Atoi(metaInfo[j:i])
	if err != nil {
		return data, err
	}
	// Find piece length.
	i = strings.Index(metaInfo, "12:piece lengthi")
	i += 16
	j = i
	for ; metaInfo[i] != 'e'; i++ {
	}
	data.pieceLength, err = strconv.Atoi(metaInfo[j:i])
	if err != nil {
		return data, err
	}
	// Find number of pieces.
	if data.length%data.pieceLength == 0 {
		data.numberOfPieces = data.length / data.pieceLength
	} else {
		data.numberOfPieces = (data.length / data.pieceLength) + 1
	}

	return data, err
}

// isCyfer returns true if character is cyfer.
func isCyfer(ch byte) bool {
	return (ch >= '0') && (ch <= '9')
}
