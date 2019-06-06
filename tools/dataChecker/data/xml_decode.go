package data

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

/*
	xml文件里不方便处理后面是空白的行，先这样吧
*/

type XmlData struct {
	Worksheet []Worksheet
}

type Worksheet struct {
	Name string `xml:"Name,attr"`
	Row  []Row  `xml:"Table>Row"`
}

type Row struct {
	Data []string `xml:"Cell>Data"`
}

func DumpIDSXml() *XmlData {
	fullName := filepath.Join(cfg.Dir.ClientProjectDir, "Assets/Data/Localization/TextRes/Languages.xml")
	if cfg.Dir.RunOnTeamCity == true {
		fullName = filepath.Join(cfg.Dir.LocalProjectDir, "Languages.xml")
	}

	content, err := ioutil.ReadFile(fullName)
	if err != nil {
		panic(err.Error())
	}

	var r XmlData
	err = xml.Unmarshal([]byte(content), &r)
	if err != nil {
		panic(err.Error())
	}

	return &r
}
