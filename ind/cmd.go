package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	// "go.bug.st/serial" // работает, но нет полного имени COM (Product)
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type COM struct {
	portName string
	port     serial.Port
}

func (com *COM) Open() (err error) {

	// Retrieve the port list
	// ports, err := serial.GetPortsList()
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
		fmt.Println("COM Open(): ERROR")
	}

	// Print the list of detected ports
	for _, p := range ports {
		fmt.Printf("Found port: %v\n", p.Product)
		if strings.Contains(p.Product, "STMicroelectronics") {
			com.portName = p.Name
		}
	}

	// Open the first serial port detected at 9600bps N81
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	com.port, err = serial.Open(com.portName, mode)
	if err != nil {
		log.Fatal(err)
		fmt.Println("cmdInd(): ERROR")
	}
	return
}

func (com *COM) Close() {
	com.port.Close()
}

func (com *COM) CmdInd(cmd string) (answer string, err error) {

	if com.port == nil || com.portName == "" {
		err = errors.New("CmdInd(): ERROR")
		fmt.Println(err)
		return
	}

	_, err = com.port.Write([]byte(cmd + "\n\r"))
	if err != nil {
		err = errors.New("CmdInd(): ERROR")
		fmt.Println(err)
	}
	// fmt.Printf("Sent %v bytes\n", n)

	// Read and print the response
	buff := make([]byte, 100)
	for { //todo ограничить время приема
		// Reads up to 100 bytes
		n, err := com.port.Read(buff)
		if err != nil {
			err = errors.New("CmdInd(): ERROR")
			fmt.Println(err)
		}
		if n == 0 {
			// fmt.Println("\nEOF")
			break
		}

		// fmt.Printf("%s", string(buff[:n]))
		answer += string(buff[:n])

		// If we receive a newline stop reading
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}

	return
}

// ----------------------------------------------------------------------------- //
//						 				Команды									 //
// ----------------------------------------------------------------------------- //

func IndOn(com COM, cmd string) (result bool, err error) { //todo переименовать
	if com.port == nil || com.portName == "" {
		err = errors.New("IndOn(): ERROR")
		fmt.Println(err)
		return
	}

	var answer string
	temp := strings.Split(cmd, "=")
	numberInd := temp[0]

	answer, err = com.CmdInd(cmd) // переименовать todo
	if err == nil && strings.Contains(answer, "OK") && strings.Contains(answer, numberInd) {
		result = true
	}
	return
}

func ButtonOn(com COM) (btn int64, err error) { // отмечать ошибку связи на форме? todo
	if com.port == nil || com.portName == "" {
		err = errors.New("ButtonOn(): ERROR")
		fmt.Println(err)
		return
	}

	var answer string
	answer, err = com.CmdInd("r70")
	if err == nil && strings.Contains(answer, "ERR") {
		btn = -1
	}
	if err == nil && strings.Contains(answer, "=") {
		temp := strings.Split(answer, "=")
		s := temp[1]
		s = strings.TrimRight(s, "\r\n")
		btn, err = strconv.ParseInt(s, 16, 64)
		// fmt.Println(btn, err)
	}
	return
}

/*func allIndsOn(com COM) (err error) {
	if com.port == nil || com.portName == "" {
		err = errors.New("allIndsOff(): ERROR")
		fmt.Println(err)
		return
	}

	_, err = com.CmdInd("w78=FF")
	_, err = com.CmdInd("w7A=FF")
	_, err = com.CmdInd("w7C=FF")
	_, err = com.CmdInd("w7E=FF")
	return
}*/

func allIndsOff(com COM) (err error) {
	if com.port == nil || com.portName == "" {
		err = errors.New("allIndsOff(): ERROR")
		fmt.Println(err)
		return
	}

	_, err = com.CmdInd("w78=0")
	_, err = com.CmdInd("w7A=0")
	_, err = com.CmdInd("w7C=0")
	_, err = com.CmdInd("w7E=0")
	return
}
