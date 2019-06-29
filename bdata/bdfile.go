package bdata

import (
	"bufio"
	"kidshelloworld.com/bindb/mod"
	"io"
	"os"
	"strconv"
	"strings"
)

func read(filepath string, seekOffset int64) ([]string, int64, error) {
	var (
		file *os.File
		err  error
	)
	if file, err = os.Open(filepath); err != nil {
		return nil, 0, err
	}
	if _, err := file.Seek(seekOffset, 0); err != nil {
		return nil, 0, err
	}

	result := make([]string, 0, 4096)
	reader := bufio.NewReader(file)
	lineNum := 0
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		lineNum += 1
		//skip header
		if seekOffset == 0 && lineNum == 1 {
			continue
		}
		result = append(result, string(line))
	}

	var (
		fileInfo os.FileInfo
	)
	if fileInfo, err = file.Stat(); err != nil {
		return nil, 0, err
	}
	return result, fileInfo.Size(), nil
}

func parse(value string) ([]mod.BinData, error) {
	values := strings.Split(value, ",")
	iinStart := values[1]
	iinEnd := values[2]
	var result []mod.BinData
	//id,iin_start,iin_end,number_length,number_luhn,scheme,brand,type,prepaid,country,
	//bank_name,bank_logo,bank_url,bank_phone,bank_city
	var (
		id                              int64
		startId, endId, currentIinStart uint32
		err                             error
	)
	if startId, err = bin2Uint32(iinStart); err != nil {
		return nil, err
	}
	endId = startId
	if "" != iinEnd {
		if endId, err = bin2Uint32(iinEnd); err != nil {
			return nil, err
		}
	}

	currentIinStart = startId
	result = make([]mod.BinData, 0, endId-startId+1)
	for {
		bindata := mod.BinData{}
		if id, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return nil, err
		}
		bindata.Id = id
		bindata.IinStart = currentIinStart
		bindata.IinEnd = currentIinStart
		bindata.NumberLength = -1
		bindata.NumberLuhn = ""
		bindata.Schema = values[5]
		bindata.Brand = values[6]
		bindata.CardType = values[7]
		bindata.Prepaid = values[8]
		bindata.Country = values[9]
		bindata.BankName = values[10]
		bindata.BankLogo = values[11]
		bindata.BankUrl = values[12]
		bindata.BankPhone = values[13]
		bindata.BankCity = values[14]
		result = append(result, bindata)
		if currentIinStart == endId {
			break
		}
		currentIinStart += 1
	}
	return result, nil
}
