package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type configType struct {
	Name        string
	ProgramName string
	Version     string
	Build       int
	Year        string
	Icon        string
	Theme       string
	ComPortName string
}

func getFyneAPP() (data configType) {
	var err error
	var data1 struct {
		Details struct {
			Icon    string
			Name    string
			Version string
			Build   int
		}
	}
	fileName := ".\\FyneAPP.toml"
	_, err = toml.DecodeFile(fileName, &data1)
	if err != nil {
		fmt.Println(err)
	}

	var data2 struct {
		Details struct {
			ProgramName string
			Year        string
			Theme       string
			ComPortName string
		}
	}
	fileName = ".\\config.toml"
	_, err = toml.DecodeFile(fileName, &data2)
	if err != nil {
		fmt.Println(err)
	}

	data.Name = data1.Details.Name
	data.ProgramName = data2.Details.ProgramName
	data.Version = data1.Details.Version
	data.Build = data1.Details.Build
	data.Year = data2.Details.Year
	data.Icon = data1.Details.Icon
	data.Theme = data2.Details.Theme
	data.ComPortName = data2.Details.ComPortName

	return
}

func writeFyneAPP(data configType) {
	f, err := os.Create(".\\config.toml")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	f.WriteString("[Details]\n")
	f.WriteString(fmt.Sprintf("ProgramName = \"%s\"\n", data.ProgramName))
	f.WriteString(fmt.Sprintf("Year = \"%s\"\n", data.Year))
	f.WriteString(fmt.Sprintf("Theme = \"%s\"\n", data.Theme))
	f.WriteString(fmt.Sprintf("COM = \"%s\"\n", data.ComPortName))
}
