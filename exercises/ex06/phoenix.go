package main

import (
	"encoding/binary"
	. "fmt"
	"log"
	"net"
	"os/exec"
	t "time"
)

var counter uint64
var port int = 9999
var ipAddr string = "127.0.0.1"
var buf = make([]byte, 16)

func spawnBackup() {
	(exec.Command("osascript", "-e", "tell app \"Terminal\" to do script \"go run Desktop/ex06/phoenix.go\"")).Run()

	Println("New backup up and running!")
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "10.22.74.201:9999") //ipAddr+":"+string(port))
	isPrimary := false
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error, something went wrong...")
	}
	// error handling?
	log.Println("Hi, I'm the backup!")

	// backup loop
	for !(isPrimary) {
		conn.SetReadDeadline(t.Now().Add(2 * t.Second))
		n, _, err := conn.ReadFromUDP(buf) // need addr for anything?
		if err != nil {
			isPrimary = true
		} else {
			counter = binary.BigEndian.Uint64(buf[:n])
		}
	}
	conn.Close()

	Println("addr: ", addr)
	spawnBackup()
	Println("I'm now the primary!")
	//addr, _ = net.ResolveUDPAddr("udp", ipAddr+":"+string(port))
	bcastConn, _ := net.DialUDP("udp", nil, addr)

	// primary loop
	for {
		Println(counter)
		counter++
		binary.BigEndian.PutUint64(buf, counter)
		//_, _ = bcastConn.WriteTo([]byte(string(buf)), addr)
		_, _ = bcastConn.Write(buf)
		t.Sleep(500 * t.Millisecond)
	}
}

// func Transmitter(port int, chans ...interface{}) {
// 	checkArgs(chans...)

// 	n := 0
// 	for range chans {
// 		n++
// 	}

// 	selectCases := make([]reflect.SelectCase, n)
// 	typeNames := make([]string, n)
// 	for i, ch := range chans {
// 		selectCases[i] = reflect.SelectCase{
// 			Dir:  reflect.SelectRecv,
// 			Chan: reflect.ValueOf(ch),
// 		}
// 		typeNames[i] = reflect.TypeOf(ch).Elem().String()
// 	}

// 	conn := conn.DialBroadcastUDP(port)
// 	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
// 	for {
// 		chosen, value, _ := reflect.Select(selectCases)
// 		buf, _ := json.Marshal(value.Interface())
// 		conn.WriteTo([]byte(typeNames[chosen]+string(buf)), addr)
// 	}
// }

// func Receiver(port int, chans ...interface{}) {
// 	checkArgs(chans...)

// 	var buf [1024]byte
// 	conn := conn.DialBroadcastUDP(port)
// 	for {
// 		n, _, _ := conn.ReadFrom(buf[0:])
// 		for _, ch := range chans {
// 			T := reflect.TypeOf(ch).Elem()
// 			typeName := T.String()
// 			if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
// 				v := reflect.New(T)
// 				json.Unmarshal(buf[len(typeName):n], v.Interface())

// 				reflect.Select([]reflect.SelectCase{{
// 					Dir:  reflect.SelectSend,
// 					Chan: reflect.ValueOf(ch),
// 					Send: reflect.Indirect(v),
// 				}})
// 			}
// 		}
// 	}
// }

// func checkArgs(chans ...interface{}) {
// 	n := 0
// 	for range chans {
// 		n++
// 	}
// 	elemTypes := make([]reflect.Type, n)

// 	for i, ch := range chans {
// 		// Must be a channel
// 		if reflect.ValueOf(ch).Kind() != reflect.Chan {
// 			panic(fmt.Sprintf(
// 				"Argument must be a channel, got '%s' instead (arg#%d)",
// 				reflect.TypeOf(ch).String(), i+1))
// 		}

// 		elemType := reflect.TypeOf(ch).Elem()

// 		// Element type must not be repeated
// 		for j, e := range elemTypes {
// 			if e == elemType {
// 				panic(fmt.Sprintf(
// 					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
// 					j+1, i+1, e.String()))
// 			}
// 		}
// 		elemTypes[i] = elemType

// 		// Element type must be encodable with JSON
// 		switch elemType.Kind() {
// 		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
// 			panic(fmt.Sprintf(
// 				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
// 				elemType.String(), i+1))
// 		case reflect.Map:
// 			if elemType.Key().Kind() != reflect.String {
// 				panic(fmt.Sprintf(
// 					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
// 					elemType.String(), i+1))
// 			}
// 		}
// 	}
// }

// func DialBroadcastUDP(port int) net.PacketConn {
// 	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
// 	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
// 	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
// 	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})

// 	f := os.NewFile(uintptr(s), "")
// 	conn, _ := net.FilePacketConn(f)
// 	f.Close()

// 	return conn
// }
