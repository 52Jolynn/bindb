package data

import (
	"bytes"
	"errors"
	"fmt"
	"git.thinkinpower.net/bindb/mod"
	"github.com/fsnotify/fsnotify"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"sync"
)

var memoryOnce sync.Once

type BinDatabase interface {
	Read(bin uint32) (mod.BinData, error)
	Save(bin uint32, binData mod.BinData) error
}

type memoryDatabase struct {

}

func (m *memory)Read(bin uint32) (mod.BinData, error)  {
	return mod.BinData{}, nil
}

func (m *memory)Save(bin uint32, bindata mod.BinData) error  {
	return nil
}

var (
	bankNameCnFileName         = "bank_name_cn.csv"
	countryCnFileName          = "country_cn.csv"
	binDataFileExt             = ".bd"
	binDataApproximateFileName = fmt.Sprintf("approximate%s", binDataFileExt)
	binDataFileName            = fmt.Sprintf("bindata%s", binDataFileExt)
	binDataHeader              = "iin_start,iin_end,number_length,number_luhn,scheme,brand,type,prepaid,country,bank_name,bank_logo,bank_url,bank_phone,bank_city"
	binDataApproximateHeader   = "iin_start,iin_end,number_length,number_luhn,scheme,brand,type,prepaid,country,bank_name,bank_logo,bank_url,bank_phone,bank_city,feedback_count"
)

var baseDataDir string
var binDataMap = make(map[uint32]*mod.BinData)
var bankNameCnMap = make(map[string]string, 3000)
var countryCnMap = make(map[string]string, 300)

func readFromFile(filepath string, append bool) {
	filename := path.Base(filepath)
	if filename == bankNameCnFileName {
		loadBankNameCnFile(filepath, append)
	} else if filename == countryCnFileName {
		loadCountryCnFile(filepath, append)
	} else if path.Ext(filepath) == binDataFileExt {
		loadBinDataFile(filepath, append)
	} else {
		logger.Infof("忽略文件: %s\n", filepath)
	}
}

func loadBankNameCnFile(filepath string, append bool) {
	if append {
		//添加
	} else {
		//全新

	}
}

func loadCountryCnFile(filepath string, append bool) {

}

func loadBinDataFile(filepath string, append bool) {

}

func CreateBankNameMapping(key, name string) error {
	if _, ok := bankNameCnMap[key]; ok {
		return nil
	}
	if baseDataDir == "" {
		return errors.New("存储地址未配置")
	}

	if err := ioutil.WriteFile(
		fmt.Sprintf("%s/%s", baseDataDir, bankNameCnFileName),
		[]byte(strings.Join([]string{key, name}, "=")), 0644); err != nil {
		logger.Error(err.Error())
		return errors.New("创建银行名称映射关系失败")
	}
	bankNameCnMap[key] = name
	return nil
}

func CreateCountryCnNameMapping(key, name string) error {
	if _, ok := countryCnMap[key]; ok {
		return nil
	}
	if baseDataDir == "" {
		return errors.New("存储地址未配置")
	}

	if err := ioutil.WriteFile(
		fmt.Sprintf("%s/%s", baseDataDir, bankNameCnFileName),
		[]byte(strings.Join([]string{key, name}, "=")), 0644); err != nil {
		logger.Error(err.Error())
		return errors.New("创建国家名称映射关系失败")
	}
	countryCnMap[key] = name
	return nil
}

func CreateBinData(bin string, bindata mod.BinData, approximate bool) error {
	var (
		uint32bin uint32
		err       error
	)
	if uint32bin, err = bin2Uint32(bin); err != nil {
		return err
	}
	if _, ok := binDataMap[uint32bin]; !ok {
		return nil
	}
	if baseDataDir == "" {
		return errors.New("存储地址未配置")
	}
	date := time.Now().Format(DatePatternCompact)
	var filepath string
	if approximate {
		filepath = strings.Join([]string{baseDataDir, date, binDataApproximateFileName}, "/")
		return saveApproximateBinData(uint32bin, filepath, bindata)
	} else {
		filepath = strings.Join([]string{baseDataDir, date, binDataFileName}, "/")
		return saveBinData(uint32bin, filepath, bindata)
	}
}

func saveApproximateBinData(bin uint32, filepath string, bindata mod.BinData) error {

}

func saveBinData(bin uint32, filepath string, bindata mod.BinData) error {
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
		strconv.FormatUint(uint64(bindata.IinStart), 10),
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

	if err := ioutil.WriteFile(filepath, data.Bytes(), 0644); err != nil {
		logger.Error(err.Error())
		return errors.New("保存bindata失败")
	}
	binDataMap[bin] = &bindata
	return nil
}

func bin2Uint32(bin string) (uint32, error) {
	var (
		ibin uint64
		err  error
	)
	if ibin, err = strconv.ParseUint(bin, 10, 32); err != nil {
		return 0, errors.New(fmt.Sprintf("invalid bin %s", bin))
	}
	return uint32(ibin), nil
}

func Query(bin string) (*mod.SimpleBinData, error) {
	var (
		uint32bin uint32
		err       error
	)
	if uint32bin, err = bin2Uint32(bin); err != nil {
		return nil, err
	}

	var (
		result *mod.BinData
		ok     bool
	)
	if result, ok = binDataMap[uint32bin]; !ok {
		return nil, errors.New(fmt.Sprintf("bin %s not found", bin))
	}

	var bankNameCn, countryCn string
	if bankNameCn, ok = bankNameCnMap[result.BankName]; !ok {
		bankNameCn = result.BankName
	}
	if countryCn, ok = countryCnMap[result.Country]; !ok {
		countryCn = result.Country
	}
	return &mod.SimpleBinData{
		BaseBinData: mod.BaseBinData{
			Schema:   result.Schema,
			Brand:    result.Brand,
			CardType: result.CardType,
			Country:  result.Country,
			BankName: result.BankName},
		BankNameCn: bankNameCn,
		CountryCn:  countryCn}, nil
}

func WatchBinDataDir(dir string) {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(error); ok {
				logger.Errorf("panic: %s\n", v.Error())
			}
			logger.Errorf("watching bin data directory error: %s\n", string(debug.Stack()))
		}
	}()
	baseDataDir = path.Dir(dir)
	loadBinDataFromDir(dir)
	beginWatching(dir)
}

func loadBinDataFromDir(dir string) {
	var (
		fileInfos []os.FileInfo
		err       error
	)
	if fileInfos, err = ioutil.ReadDir(dir); err != nil {
		logger.Errorf("load bin data from %s failed! error: %s\n", dir, err.Error())
	} else {
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				loadBinDataFromDir(fileInfo.Name())
			} else {
				readFromFile(fileInfo.Name(), false)
			}
		}
	}
}

func beginWatching(dir string) {
	var (
		watcher *fsnotify.Watcher
		err     error
	)
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		logger.Error(err)
	}
	defer func() {
		if err = watcher.Close(); err != nil {
			logger.Error(err)
		}
	}()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.Infof("file modified %s:\n", event.Name)
					readFromFile(event.Name, true)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					logger.Infof("file created %s:\n", event.Name)
					readFromFile(event.Name, false)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("watch %s error:%s\n", dir, err.Error())
			}
		}
	}()

	if err = watcher.Add(dir); err != nil {
		logger.Error(err)
	}
	<-done
}
