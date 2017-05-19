package main

import (
	"encoding/json"
	"errors"
)

type RegexpConfig struct {
	Name string `json:"name"`
	Re   string `json:"re"`
}

type Config struct {
	Regexps		 []RegexpConfig `json:"regexps"`
	Interval	 int            `json:"interval"`
	Interface	 string         `json:"interface"`
	IpAddress	 string         `json:"ip"`
	Port		 int            `json:"port"`
	OutputFile	 string         `json:"output_file"`
	SortBy		 string		`json:"sortby"`
}

func NewConfig(config_data []byte) (config Config, err error) {
	config = Config{
		Regexps:          []RegexpConfig{},
		Interval:         3,
		Interface:        "any",
		IpAddress:        "",
		Port:             11211,
		OutputFile:	  "",
		SortBy:		  "rcount",
	}
	if len(config_data) > 0 {
		err = json.Unmarshal(config_data, &config)
		if err != nil {
			return config, err
		}
	}
	for _, re := range config.Regexps {
		if re.Name == "" || re.Re == "" {
			return config, errors.New(
				"Config error: regular expressions must have both a 're' and 'name' field.")
		}
	}

	return config, nil
}
