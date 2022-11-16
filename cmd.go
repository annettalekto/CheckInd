package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// COM вот
type COM struct {
	portName  string
	port      serial.Port
	available []string
	err       error
}

// Open открыть COM вот
func (com *COM) Open() error {

	// пробуем открыть COM сохранееный в config
	if "" != config.ComPortName {
		if err := com.OpenOne(config.ComPortName); err == nil {
			return nil
		}
	}

	// Retrieve the port list
	// ports, err := serial.GetPortsList()
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		fmt.Printf("COM GetDetailedPortsList(): %v\n", err)
		com.err = errors.New("Ошибка COM: не получен список доступный COM-портов")
		return com.err
	}
	if len(ports) == 0 {
		fmt.Println("Ошибка COM: не найден ни один COM-порт")
		com.err = errors.New("Ошибка COM: не найден ни один COM-порт")
		return com.err
	}

	// ищем наш адаптер
	var temp []string
	for _, p := range ports {
		fmt.Printf("Found port: %v\n", p.Product)
		temp = append(temp, p.Name)
		if strings.Contains(p.Product, "STMicroelectronics") {
			com.portName = p.Name
		}
	}
	com.available = temp
	if com.portName == "" {
		fmt.Printf("Ошибка COM: не найден нужный COM-порт")
		com.err = errors.New("Ошибка COM: не найден нужный COM-порт")
		return com.err
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
		fmt.Printf("COM Open(): %v\n", err)
		com.err = errors.New("Ошибка COM: ошибка открытия COM-порта")
	} else {
		com.err = nil
		config.ComPortName = com.portName
		writeFyneAPP(config)
	}
	return com.err
}

// OpenOne открыть COM вот
func (com *COM) OpenOne(sCOM string) error {
	var err error

	// Open the first serial port detected at 9600bps N81
	mode := &serial.Mode{
		// BaudRate: 19200,
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	com.portName = sCOM
	com.port, err = serial.Open(com.portName, mode)
	if err != nil {
		fmt.Printf("COM Open(): %v\n", err)
		com.err = errors.New("Ошибка COM: ошибка открытия COM-порта")
	} else {
		com.err = nil
	}
	return com.err
}

// PortSearch найти порты
func (com *COM) PortSearch() error {
	// Retrieve the port list
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		fmt.Printf("COM GetDetailedPortsList(): %v\n", err)
		com.err = errors.New("Ошибка COM: не получен список доступный COM-портов")
		com.available = nil
		return com.err
	}
	if len(ports) == 0 {
		fmt.Println("Ошибка COM: не найден ни один COM-порт")
		com.err = errors.New("Ошибка COM: не найден ни один COM-порт")
		com.available = nil
		return com.err
	}

	// Print the list of detected ports
	var temp []string
	for _, p := range ports {
		fmt.Printf("Found port: %v\n", p.Product)
		temp = append(temp, p.Name)
	}
	com.available = temp

	return com.err
}

// Close закрыть COM
func (com *COM) Close() {
	if nil == com.port {
		return
	}
	com.port.Close()
}

// Cmd отправить команду в COM
func (com *COM) Cmd(cmd string) (string, error) {
	if nil != com.err || nil == com.port || "" == com.portName {
		return "", errors.New("Ошибка COM")
	}
	var answer string

	if _, err := com.port.Write([]byte(cmd + "\n\r")); err != nil {
		com.err = errors.New("Ошибка COM: ошибка записи данных")
		fmt.Println(com.err)
		return "", com.err
	}

	// Read and print the response
	buff := make([]byte, 100)
	start := time.Now()
	for time.Since(start) <= (time.Second / 2) {
		// Reads up to 100 bytes
		n, err := com.port.Read(buff)
		if err != nil {
			com.err = errors.New("Ошибка COM: ошибка чтения данных")
			fmt.Println(com.err)
		}
		if n == 0 {
			break
		}

		answer += string(buff[:n])

		// If we receive a newline stop reading
		if strings.Contains(string(buff[:n]), "\n") {
			break
		}
	}

	fmt.Printf("cmd: %s, answer: %s\n", cmd, answer)
	return answer, com.err
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

	if answer, err = com.Cmd(cmd); err != nil {
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

	if answer, err = com.Cmd("r70"); err != nil {
		return
	}
	if strings.Contains(answer, "r70") && strings.Contains(answer, "\r\n") { // есть начало и конец строки
		if strings.Contains(answer, "ERR") {
			btn = -1
		} else if strings.Contains(answer, "=") {
			temp := strings.Split(answer, "=")
			s := temp[1]
			s = strings.TrimRight(s, "\r\n")
			btn, err = strconv.ParseInt(s, 16, 64)
		} else {
			err = errors.New("некорректый ответ")
		}
	} else {
		err = errors.New("некорректный ответ")
	}

	return
}

// CheckRelay проверить какие биты (реле) установлены в единицу
// bits - биты которые нужно установить
// relay - возвращает установленные биты
func (com *COM) CheckRelay(bits int) (setbits int64, err error) {
	var answer string

	// установить все нужные биты, например: w42=FF (первые биты лишние, таких реле нет на плате)
	// прочитать установленные в единицу биты: r45=0F (вернулись те 4, что есть на плате)

	cmd := "w42=" + fmt.Sprintf("%X", bits)
	//answer, err =
	com.Cmd(cmd)
	time.Sleep(20 * time.Millisecond)

	answer, err = com.Cmd("r45")
	if err != nil {
		return
	}

	if strings.Contains(answer, "r45") && strings.Contains(answer, "\r\n") {
		if strings.Contains(answer, "ERR") {
			setbits = -1
		} else if strings.Contains(answer, "=") {
			temp := strings.Split(answer, "=") // r45=0F
			s := temp[1]
			s = strings.TrimRight(s, "\r\n")
			setbits, err = strconv.ParseInt(s, 16, 64)
		} else {
			err = errors.New("некорректый ответ")
		}
	} else {
		err = errors.New("некорректный ответ")
	}

	return
}

// IndsOff погасить все индикаторы
func (com *COM) IndsOff() (err error) {

	_, err = com.Cmd("w78=0")
	_, err = com.Cmd("w7A=0")
	_, err = com.Cmd("w7C=0")
	_, err = com.Cmd("w7E=0")
	return
}
