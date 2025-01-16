package scanning

import (
	"encoding/base64"
	"encoding/json"
)

const (
	Version = iota
	V1
	V2
)

type Scan struct {
	Ip          string      `json:"ip"`
	Port        uint32      `json:"port"`
	Service     string      `json:"service"`
	Timestamp   int64       `json:"timestamp"`
	DataVersion int         `json:"data_version"`
	Data        interface{} `json:"data"`
}

type V1Data struct {
	ResponseBytesUtf8 []byte `json:"response_bytes_utf8"`
}

type V2Data struct {
	ResponseStr string `json:"response_str"`
}

type ServiceData struct {
	Ip          string      `json:"ip"`
	Port        uint32      `json:"port"`
	Service     string      `json:"service"`
	Timestamp   int64       `json:"timestamp"`
	Response	string		`json:"response"`
}

func (sd *ServiceData) UnmarshalJSON(data []byte) error {
	var scan Scan
	err := json.Unmarshal(data, &scan)
	if err != nil {
		return err
	}

	sd.Ip, sd.Port, sd.Service, sd.Timestamp = scan.Ip, scan.Port, scan.Service, scan.Timestamp
	if scan.DataVersion == V1 {
		str := scan.Data.(map[string]interface{})["response_bytes_utf8"].(string)
		decodedBytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}
		sd.Response = string(decodedBytes)
	} else {
		str := scan.Data.(map[string]interface{})["response_str"].(string)
		sd.Response = str
	}
	return nil
}
