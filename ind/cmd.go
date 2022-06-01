package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	// "go.bug.st/serial" // работает, но нет полного имени COM (Product)
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// COM вот
type COM struct {
	process  bool
	portName string
	port     serial.Port
	err      error
}

// Open открыть COM вот
func (com *COM) Open() (err error) {
	// Retrieve the port list
	// ports, err := serial.GetPortsList()
	ports, e := enumerator.GetDetailedPortsList()
	if e != nil {
		fmt.Println("COM GetDetailedPortsList(): ERROR")
		com.err = e
		return e
	}
	if len(ports) == 0 {
		fmt.Println("COM: Не найден ни один COM-порт")
		com.err = errors.New("COM: Не найден ни один COM-порт")
		return
	}

	// Print the list of detected ports
	for _, p := range ports {
		fmt.Printf("Found port: %v\n", p.Product)
		if strings.Contains(p.Product, "STMicroelectronics") {
			com.portName = p.Name
		}
	}
	if com.portName == "" {
		return errors.New("COM: Не найден нужный COM-порт")
	}

	// Open the first serial port detected at 9600bps N81
	mode := &serial.Mode{
		// BaudRate: 19200,
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	com.port, err = serial.Open(com.portName, mode)
	if err != nil {
		// log.Fatal(err)
		fmt.Printf("COM Open(): %e\n", err)
		com.err = err
	}
	return
}

// Close закрыть COM
func (com *COM) Close() {
	if com.err != nil || com.port == nil || com.portName == "" {
		return
	}
	com.port.Close()
}

// Cmd отправить команду в COM
func (com *COM) Cmd(cmd string) (answer string, err error) {
	if com.err != nil || com.port == nil || com.portName == "" {
		err = errors.New("com == nil")
		return
	}

	if !com.process {
		com.process = true

		_, err = com.port.Write([]byte(cmd + "\n\r"))
		if err != nil {
			err = errors.New("ошибка записи")
			fmt.Println(err)
			com.err = err
			return
		}
		// fmt.Printf("Sent %v bytes\n", n)

		// Read and print the response
		buff := make([]byte, 100)
		start := time.Now()
		for time.Since(start) <= (time.Second / 2) { //todo ограничить время приема
			// Reads up to 100 bytes
			n, err := com.port.Read(buff)
			if err != nil {
				err = errors.New("ошибка чтения")
				com.err = err
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
		fmt.Println(answer)

		com.process = false
	} else {
		err = errors.New("process")
		fmt.Println("ERR PROCESS")
	}

	return
}

// ----------------------------------------------------------------------------- //
//						 				Команды									 //
// ----------------------------------------------------------------------------- //

// CheckInd проверить индикатор
// cmd должен указать индикатор и сегменты, которые нужно "зажечь" на плате: w78=01
func (com *COM) CheckInd(cmd string) (result bool, err error) {

	var answer string
	temp := strings.Split(cmd, "=")
	numberInd := temp[0] // например "w7E"

	answer, err = com.Cmd(cmd) // переименовать todo
	// fmt.Println(answer)

	if err != nil {
		err = errors.New("ошибка передачи данных")
		return
	}
	if strings.Contains(answer, numberInd) && strings.Contains(answer, "\r\n") { // есть начало и конец строки
		if strings.Contains(answer, "OK") {
			result = true
		} else if strings.Contains(answer, "ERR") {
			result = false
		} else {
			err = errors.New("некорректый ответ")
		}
	} else {
		err = errors.New("некорректный ответ")
	}

	return
}

// CheckButton проверить какие из кнопок нажаты
// btn - возвращает нажатые кнопки
func (com *COM) CheckButton() (btn int64, err error) {
	var answer string

	answer, err = com.Cmd("r70")
	// fmt.Println(answer)

	if err != nil {
		err = errors.New("ошибка передачи данных")
		return
	}

	if strings.Contains(answer, "r70") && strings.Contains(answer, "\r\n") { // есть начало и конец строки
		if strings.Contains(answer, "ERR") {
			btn = -1 //
		} else if strings.Contains(answer, "=") {
			temp := strings.Split(answer, "=")
			s := temp[1]
			s = strings.TrimRight(s, "\r\n")
			btn, err = strconv.ParseInt(s, 16, 64)
		} else {
			err = errors.New("некорректынф ответ")
		}
	} else {
		err = errors.New("некорректный ответ")
	}

	return
}

// IndsOff погасить все индикаторы
func (com COM) IndsOff() (err error) {

	_, err = com.Cmd("w78=0")
	_, err = com.Cmd("w7A=0")
	_, err = com.Cmd("w7C=0")
	_, err = com.Cmd("w7E=0")
	return
}
