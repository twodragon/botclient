package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hero/utils"
	"log"
	"net"
	"strings"
	"time"
)

type Socket struct {
	Conn net.Conn
}

var (
	username string
	userlen  int
	cid      uint64
	Location *utils.Location
	Target   *utils.Location
)

func main() {
	// Bağlanılacak sunucunun IP adresi ve portu karakter id username ve parolası
	serverAddr := "188.132.128.230:4515"
	cid = 1044
	username = "id"
	password := "şifre"

	Location = ConvertPointToLocation("(313.0,161.0)")
	Target = ConvertPointToLocation("(313.0,163.0)")
	// Sunucuya bağlan
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Sunucuya bağlanırken hata oluştu:", err)
		return
	}

	s := &Socket{Conn: conn}
	//var s *Socket
	// İlk paketi gönder
	pkt := utils.Packet{0xAA, 0x55, 0x01, 0x00, 0x38, 0x55, 0xAA}
	s.sendbyte(pkt)

	// İkinci paketi gönder

	login := utils.Packet{0xaa, 0x55, 0xff, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x40, 0x00, 0x00, 0x55, 0xaa}
	userlen = len(username)
	login.Insert([]byte{byte(userlen)}, 9)
	login.Insert([]byte(username), 10)
	data := []byte(password)
	hashx := sha256.Sum256(data)
	password = string(hashx[:])
	password = fmt.Sprintf("%X", password)

	login.Insert([]byte(password), userlen+11)
	s.sendbyte(login)
	for {
		s.handleResponse(conn)
	}

}

func sendstring(conn net.Conn, packet string) {
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

func (s *Socket) sendbyte(packet []byte) {
	if _, err := s.Conn.Write(packet); err != nil {
		log.Printf("Packet send error: %s", err)
		return
	}
	fmt.Printf("Paket gönderildi: %X\n", packet)
	time.Sleep(time.Millisecond * 200)
}

func ConvertPointToLocation(point string) *utils.Location {

	location := &utils.Location{}
	parts := strings.Split(strings.Trim(point, "()"), ",")
	if parts[0] != "" && parts[1] != "" {
		location.X = utils.StringToFloat64(parts[0])
		location.Y = utils.StringToFloat64(parts[1])
	} else {
		location.X = 0
		location.Y = 0
	}
	return location
}

// Sunucudan gelen cevabı işler
func (s *Socket) handleResponse(conn net.Conn) {
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	//response := recognizePacket(buffer[:n])
	response, err := s.recognizePacket(buffer[:n])
	if err != nil {
		log.Println("recognize packet error:", err)
	}
	if len(response) < 0 {
		return
	}
	//response := hex.EncodeToString(buffer[:n])
	//log.Printf("Received response: %s", response)
}

func (s *Socket) recognizePacket(data []byte) ([]byte, error) { //test read pkg example with chat types
	pkgType := utils.BytesToInt(data[4:6], false)
	if pkgType == 19714 || pkgType == 8706 || pkgType == 8705 || pkgType == 22809 {
		return nil, nil
	} else {
		//log.Printf("%d", pkgType)
	}
	//https://go.dev/play/p/fvUmlP6jcnq tool link
	switch pkgType {

	case 1:
		w8 := utils.Packet{0xaa, 0x55, 0x02, 0x00, 0x00, 0x02, 0x55, 0xaa}
		s.sendbyte(w8)
	case 3:
		serversec := utils.Packet{0xaa, 0x55, 0x06, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
		s.sendbyte(serversec)
	case 5:
		usersec := utils.Packet{0xaa, 0x55, 0x0e, 0x00, 0x01, 0x01, 0x01, 0x30, 0x3f, 0x87, 0x01, 0x00, 0x55, 0xaa}
		usersec.Insert([]byte{byte(userlen)}, 6)
		usersec.Insert([]byte(username), 7)
		s.sendbyte(usersec)
	case 258:
		karaktersec := utils.Packet{0xaa, 0x55, 0x0d, 0x00, 0x01, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
		karaktersec.Insert(utils.IntToBytes(cid, 4, true), 6)
		s.sendbyte(karaktersec)
	case 261:
		start := utils.Packet{0xaa, 0x55, 0x03, 0x00, 0x62, 0x02, 0x00, 0x55, 0xaa} //oyunun baslaması için gerekli
		s.sendbyte(start)
	case 8705: //yanına gelen karakterlerin movementi
	case 8706: //yanına gelen karakterlerin movementi
	case 29185:
		movement := utils.Packet{0xaa, 0x55, 0x21, 0x00, 0x22, 0x01,
			0x00, 0x9c, 0x81, 0x40,
			0x00, 0x9c, 0x81, 0x40, 0x5a, 0x00, 0x00, 0x00, 0xa0, 0x40, 0x00, 0x55, 0xaa}

		coordinate := &utils.Location{X: Location.X, Y: Location.Y}
		target := &utils.Location{X: Target.X, Y: Target.Y}

		movement.Insert(utils.FloatToBytes(coordinate.X, 4, true), 6)  // coordinate-x
		movement.Insert(utils.FloatToBytes(coordinate.Y, 4, true), 10) // coordinate-y
		movement.Insert(utils.FloatToBytes(target.X, 4, true), 18)     // target-x
		movement.Insert(utils.FloatToBytes(target.Y, 4, true), 22)     // target-y
		s.sendbyte(movement)
		ping := utils.Packet{0xaa, 0x55, 0x01, 0x00, 0x18, 0x55, 0xaa}
		s.sendbyte(ping)

		/*coordinate := &utils.Location{X: utils.BytesToFloat(data[6:10], true), Y: utils.BytesToFloat(data[10:14], true)}
		target := &utils.Location{X: utils.BytesToFloat(data[18:22], true), Y: utils.BytesToFloat(data[22:26], true)}*/
		/*cordgo := "aa552100220100809c43000a2143009c814000809c4300002343009c81405a000000a0400055aa"
		sendstring(conn, cordgo)*/

		silahcikart := utils.Packet{0xaa, 0x55, 0x02, 0x00, 0x43, 0x01, 0x55, 0xaa} //silahı ele al
		s.sendbyte(silahcikart)
	case 28929: // normal chat
		charlen := utils.BytesToInt([]byte{data[8]}, true)
		charname := string(data[8 : 9+charlen])
		messageLen := utils.BytesToInt([]byte{data[9+charlen]}, true)
		message := string(data[10+charlen : 11+charlen+messageLen])
		log.Printf(fmt.Sprintf(" %s: %s \n", charname, message))

	case 28930: // whisper chat
		charlen := utils.BytesToInt([]byte{data[8]}, true)
		charname := string(data[8 : 9+charlen])
		messageLen := utils.BytesToInt([]byte{data[9+charlen]}, true)
		message := string(data[10+charlen : 11+charlen+messageLen])
		log.Printf(fmt.Sprintf("! %s: %s \n", charname, message))

	case 28933: // roar chat
		charlen := utils.BytesToInt([]byte{data[8]}, true)
		charname := string(data[8 : 9+charlen])
		messageLen := utils.BytesToInt([]byte{data[9+charlen]}, true)
		message := string(data[10+charlen : 11+charlen+messageLen])
		log.Printf(fmt.Sprintf("# %s: %s \n", charname, message))

	case 28946: // valorus chat
		charlen := utils.BytesToInt([]byte{data[6]}, true)
		charname := string(data[6 : 7+charlen])
		messageLen := utils.BytesToInt([]byte{data[7+charlen]}, true)
		message := string(data[8+charlen : 9+charlen+messageLen])
		log.Printf(fmt.Sprintf("' %s: %s \n", charname, message))
		//log.Printf("\n % X", data)

	case 28942: //shout
		charlen := utils.BytesToInt([]byte{data[6]}, true)
		charname := string(data[6 : 7+charlen])
		messageLen := utils.BytesToInt([]byte{data[7+charlen]}, true)
		message := string(data[8+charlen : 9+charlen+messageLen])
		log.Printf(" /Shout [%s]:%s\n", charname, message)
	}
	return nil, nil
}
