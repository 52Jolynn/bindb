package data

import (
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
)

var (
	bankNameCnFile = "bank_name_cn.csv"
	countryCnFile  = "country_cn.csv"
	binDataFileExt = ".bd"
)

var binDataMap map[uint32]*mod.BinData
var bankNameCnMap map[string]string
var countryCnMap map[string]string

func readFromFile(filepath string, append bool) {
	filename := path.Base(filepath)
	if filename == bankNameCnFile {
		loadBankNameCnFile(filepath, append)
	} else if filename == countryCnFile {
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

func CreateBankNameMapping(key, name string) {

}

func CreateCountryCnNameMapping(key, name string) {

}

func CreateBinData(bin string, bindata mod.BinData) {

}

func Query(bin string) (*mod.SimpleBinData, error) {
	var (
		ibin uint64
		err  error
	)
	if ibin, err = strconv.ParseUint(bin, 10, 32); err != nil {
		return nil, errors.New(fmt.Sprintf("invalid bin %s", bin))
	}

	var (
		result *mod.BinData
		ok     bool
	)
	if result, ok = binDataMap[uint32(ibin)]; !ok {
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
