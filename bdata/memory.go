package bdata

import (
	"bytes"
	"errors"
	"fmt"
	"git.thinkinpower.net/bindb/file"
	"git.thinkinpower.net/bindb/mod"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	memoryCreateOnce sync.Once
	memory           *memoryDatabase
)

var (
	binDataFileExt             = ".bd"
	binDataApproximateFileExt  = ".bd2"
	binDataApproximateFileName = fmt.Sprintf("approximate%s", binDataApproximateFileExt)
	binDataFileName            = fmt.Sprintf("bindata%s", binDataFileExt)
	binDataHeader              = "id,iin_start,iin_end,number_length,number_luhn,scheme,brand,type,prepaid,country,bank_name,bank_logo,bank_url,bank_phone,bank_city"
	binDataMappingFile         = binDataFileHandler{bytesMap: make(map[string]int64), reloading: make(chan file.FileEvent)}
)

type binDataFileHandler struct {
	bytesMap  map[string]int64
	reloading chan file.FileEvent
}

type memoryDatabase struct {
	dataMap        map[uint32]mod.BinData
	approximateMap map[uint32]map[int64]mod.BinData
	dataDir        string
}

func NewMemoryDatabase() BinDatabase {
	memoryCreateOnce.Do(func() {
		memory = &memoryDatabase{dataMap: make(map[uint32]mod.BinData), approximateMap: make(map[uint32]map[int64]mod.BinData)}
	})
	return memory
}

func (m *memoryDatabase) Init(cfg BinDataConfig) error {
	var (
		filepaths []string
		err       error
	)
	if filepaths, err = file.SearchDir(cfg.DataDir, func(filepath string) bool {
		ext := path.Ext(filepath)
		return binDataFileExt == ext || binDataApproximateFileExt == ext
	}); err != nil {
		logger.Errorf("memory database init failed, error: %s, dataDir: %s", err, cfg.DataDir)
	}

	var (
		filedata []string
		filesize int64
	)
	initFailure := errors.New("初始化内存数据库失败")
	for _, filepath := range filepaths {
		if filedata, filesize, err = read(filepath, 0); err != nil {
			logger.Errorf("读取文件失败, error: %s, filepath: %s", err, filepath)
			return initFailure
		}

		ext := path.Ext(filepath)
		approximate := false
		if ext == binDataApproximateFileExt {
			approximate = true
		}
		for _, value := range filedata {
			var data []mod.BinData
			if data, err = parse(value); err != nil {
				logger.Errorf("parse bin data error: %s, data: %s", err, value)
			}
			for _, d := range data {
				m.save2Memory(d.IinStart, d, approximate)
			}
		}
		if size, ok := binDataMappingFile.bytesMap[filepath]; ok {
			binDataMappingFile.bytesMap[filepath] = size + filesize
		} else {
			binDataMappingFile.bytesMap[filepath] = filesize
		}
	}

	memory.dataDir = cfg.DataDir
	go registerBinDataRefresher()
	AddFileListener(func(event file.FileEvent) {
		binDataMappingFile.reloading <- event
	})
	return nil
}

func registerBinDataRefresher() {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(error); ok {
				logger.Errorf("panic: %s\n", v.Error())
			}
			logger.Errorf("refresh bin data file error: %s", string(debug.Stack()))
		}
	}()
	for {
		select {
		case event := <-binDataMappingFile.reloading:
			refreshBinData(event)
		}
	}
}

func (m *memoryDatabase) ReadExact(bin uint32) (mod.BinData, error) {
	if result, ok := m.dataMap[bin]; ok {
		return result, nil
	}
	return NullBinData, errors.New(fmt.Sprintf("%d not found", bin))
}

func (m *memoryDatabase) ReadApproximate(bin uint32) ([]mod.BinData, error) {
	if value, ok := m.approximateMap[bin]; ok {
		result := make([]mod.BinData, len(value))
		for _, v := range result {
			result = append(result, v)
		}
		return result, nil
	}
	return nil, errors.New(fmt.Sprintf("%d not found", bin))
}

func (m *memoryDatabase) Save(bin uint32, bindata mod.BinData, approximate bool) error {
	m.save2Memory(bin, bindata, approximate)
	if m.dataDir == "" {
		return errors.New("存储地址未配置")
	}

	date := time.Now().Format(DatePatternCompact)
	var filepath string
	if approximate {
		filepath = strings.Join([]string{m.dataDir, date, binDataApproximateFileName}, "/")
	} else {
		filepath = strings.Join([]string{m.dataDir, date, binDataFileName}, "/")
	}
	return write2File(bin, filepath, bindata)
}

func (m *memoryDatabase) save2Memory(bin uint32, bindata mod.BinData, approximate bool) {
	if approximate {
		var (
			valueMap map[int64]mod.BinData
			ok       bool
		)
		if valueMap, ok = m.approximateMap[bin]; !ok {
			valueMap = make(map[int64]mod.BinData, 10)
		} else {
			if _, ok = valueMap[bindata.Id]; !ok {
				valueMap[bindata.Id] = bindata
			}
		}
		m.approximateMap[bin] = valueMap
	} else {
		if _, ok := m.dataMap[bin]; !ok {
			m.dataMap[bin] = bindata
		}
	}
}

func refreshBinData(e file.FileEvent) {
	filepath := e.Filepath
	var seekOffset int64 = 0
	if !e.FileCreated {
		seekOffset = binDataMappingFile.bytesMap[filepath]
	}

	var (
		filedata []string
		filesize int64
		err      error
	)
	if filedata, filesize, err = read(filepath, seekOffset); err != nil {
		logger.Errorf("read bin data error: %s", err)
		return
	}

	ext := path.Ext(filepath)
	approximate := false
	if ext == binDataApproximateFileExt {
		approximate = true
	}

	for _, fd := range filedata {
		var binDataSet []mod.BinData
		if binDataSet, err = parse(fd); err != nil {
			logger.Errorf("parse bin binDataSet error: %s, binDataSet: %s", err, fd)
			continue
		}
		for _, d := range binDataSet {
			memory.save2Memory(d.IinStart, d, approximate)
		}
	}
	binDataMappingFile.bytesMap[filepath] += filesize
}

func write2File(bin uint32, filepath string, bindata mod.BinData) error {
	data := bytes.Buffer{}
	if _, err := os.Stat(filepath); err != nil && os.IsNotExist(err) {
		//文件不存在, 需要写入header
		data.Write([]byte(binDataHeader))
	}
	//写入数据
	iinEnd := ""
	if bindata.IinEnd != 0 {
		iinEnd = strconv.FormatUint(uint64(bindata.IinEnd), 10)
	}
	numLen := ""
	if bindata.NumberLength == 0 {
		numLen = strconv.FormatUint(uint64(bindata.NumberLength), 10)
	}
	data.WriteString(strings.Join([]string{
		strconv.FormatUint(uint64(bin), 10),
		iinEnd,
		numLen,
		bindata.NumberLuhn,
		bindata.Schema,
		bindata.CardType,
		bindata.Prepaid,
		bindata.Country,
		bindata.BankName,
		bindata.BankLogo,
		bindata.BankUrl,
		bindata.BankPhone,
		bindata.BankCity}, ","))

	if err := ioutil.WriteFile(filepath, data.Bytes(), os.ModeAppend); err != nil {
		logger.Error(err.Error())
		return errors.New("保存bindata失败")
	}
	return nil
}
