package mister

import (
	"github.com/wizzomafizzo/mrext/pkg/config"
	"gopkg.in/ini.v1"
)

const downloaderIniFile = config.SdFolder + "/downloader.ini"

type DownloaderIni struct {
	IniFile *ini.File
	Dbs     map[string]string
}

func LoadDownloaderIni() (*DownloaderIni, error) {
	iniFile, err := ini.Load(downloaderIniFile)
	if err != nil {
		return nil, err
	}

	dbs := make(map[string]string)
	for _, section := range iniFile.Sections() {
		if section.HasKey("db_url") {
			dbs[section.Name()] = section.Key("db_url").String()
		}
	}

	return &DownloaderIni{
		IniFile: iniFile,
		Dbs:     dbs,
	}, nil
}

func (d *DownloaderIni) Save() error {
	return d.IniFile.SaveTo(downloaderIniFile)
}

func (d *DownloaderIni) AddDb(name, url string) error {
	section, err := d.IniFile.NewSection(name)
	if err != nil {
		return err
	}

	section.Key("db_url").SetValue(url)
	d.Dbs[name] = url

	return nil
}

func (d *DownloaderIni) RemoveDb(name string) error {
	delete(d.Dbs, name)
	d.IniFile.DeleteSection(name)
	return nil
}

func (d *DownloaderIni) HasDb(name string) bool {
	_, ok := d.Dbs[name]
	return ok
}
