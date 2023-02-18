package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hero/utils"
	"log"
	"net"
	"time"
)

func main() {
	// Bağlanılacak sunucunun IP adresi ve portu karakter id username ve parolası
	serverAddr := "127.0.0.1:4515"
	cid := uint64(20)
	username := "admin"
	password := "zz"
	// Sunucuya bağlan
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Sunucuya bağlanırken hata oluştu:", err)
		return
	}

	// İlk paketi gönder
	pkt := utils.Packet{0xAA, 0x55, 0x01, 0x00, 0x38, 0x55, 0xAA}
	sendbyte(conn, pkt)

	// İkinci paketi gönder

	login := utils.Packet{0xaa, 0x55, 0xff, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x40, 0x00, 0x00, 0x55, 0xaa}
	userlen := len(username)
	login.Insert([]byte{byte(userlen)}, 9)
	login.Insert([]byte(username), 10)
	data := []byte(password)
	hashx := sha256.Sum256(data)
	password = string(hashx[:])
	password = fmt.Sprintf("%X", password)

	login.Insert([]byte(password), userlen+11)
	sendbyte(conn, login)
	w8 := utils.Packet{0xaa, 0x55, 0x02, 0x00, 0x00, 0x02, 0x55, 0xaa}
	sendbyte(conn, w8)

	serversec := utils.Packet{0xaa, 0x55, 0x06, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
	sendbyte(conn, serversec)

	usersec := utils.Packet{0xaa, 0x55, 0x0e, 0x00, 0x01, 0x01, 0x01, 0x30, 0x3f, 0x87, 0x01, 0x00, 0x55, 0xaa}
	usersec.Insert([]byte{byte(userlen)}, 6)
	usersec.Insert([]byte(username), 7)
	sendbyte(conn, usersec)

	ping := utils.Packet{0xaa, 0x55, 0x01, 0x00, 0x18, 0x55, 0xaa}
	sendbyte(conn, ping)

	karaktersec := utils.Packet{0xaa, 0x55, 0x0d, 0x00, 0x01, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
	karaktersec.Insert(utils.IntToBytes(cid, 4, true), 6)
	sendbyte(conn, karaktersec)

	start := utils.Packet{0xaa, 0x55, 0x03, 0x00, 0x62, 0x02, 0x00, 0x55, 0xaa} //oyunun baslaması için gerekli
	sendbyte(conn, start)

	cordgo := "aa552100220100809c43000a2143009c814000809c4300002343009c81405a000000a0400055aa"
	sendPacket(conn, cordgo)
	silahcikart := utils.Packet{0xaa, 0x55, 0x02, 0x00, 0x43, 0x01, 0x55, 0xaa} //silahı ele al
	sendPacket(conn, string(silahcikart))

	for {
		handleResponse(conn)
	}

}

func sendPacket(conn net.Conn, packet string) {
	data, err := hex.DecodeString(packet)
	if err != nil {
		fmt.Println("Paket dönüştürülürken hata oluştu:", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Paket gönderilirken hata oluştu:", err)
		return
	}
	fmt.Printf("Paket gönderildi: %s\n", packet)
	time.Sleep(time.Millisecond * 200)
}

func sendbyte(conn net.Conn, packet []byte) {
	if _, err := conn.Write(packet); err != nil {
		log.Printf("Packet send error: %s", err)
		return
	}
	fmt.Printf("Paket gönderildi: %X\n", packet)
	time.Sleep(time.Millisecond * 200)
}

// Sunucudan gelen cevabı işler
func handleResponse(conn net.Conn) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	response := hex.EncodeToString(buffer[:n])
	log.Printf("Received response: %s", response)
}
