package bdata

import (
	"bufio"
	"errors"
	"fmt"
	"git.thinkinpower.net/bindb/file"
	"git.thinkinpower.net/bindb/mod"
	"github.com/fsnotify/fsnotify"
	logger "github.com/sirupsen/logrus"
	"io"
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
	bankNameCnFileName = "bank_name_cn.csv"
	countryCnFileName  = "country_cn.csv"
)

var (
	config                 = BinDataConfig{}
	currentBinDatabase     BinDatabase
	bankNameCnMapping      = mappingFile{bytes: 0, reloading: make(chan file.FileEvent), handler: refreshBankName, dataMap: make(map[string]string, 3000)}
	countryCnMapping       = mappingFile{bytes: 0, reloading: make(chan file.FileEvent), handler: refreshCountry, dataMap: make(map[string]string, 300)}
	NullBinData            mod.BinData
	setBinDatabaseModeOnce sync.Once
)

type fileEventListener func(file.FileEvent)

var fileEventListenerList []fileEventListener

type BinDataConfig struct {
	DataDir string
}

type mappingFile struct {
	bytes     int64
	reloading chan file.FileEvent
	handler   func(file.FileEvent)
	dataMap   map[string]string
}

type BinDatabase interface {
	Init(cfg BinDataConfig) error
	ReadExact(bin uint32) (mod.BinData, error)
	ReadApproximate(bin uint32) ([]mod.BinData, error)
	Save(bin uint32, binData mod.BinData, approximate bool) error
}

func SetBinDatabaseMode(mode string) {
	setBinDatabaseModeOnce.Do(func() {
		if BinDatabaseModeMemory == mode {
			currentBinDatabase = NewMemoryDatabase()
		} else if BinDatabaseModeRedis == mode {
			//need to be impl
		} else {
			currentBinDatabase = NewMemoryDatabase()
		}
		if err := currentBinDatabase.Init(config); err != nil {
			logger.Fatal("refresh bin data error: %s", err.Error())
		}
	})
}

func AddFileListener(listener fileEventListener) {
	fileEventListenerList = append(fileEventListenerList, listener)
}

func CreateBankNameMapping(key, name string) error {
	if _, ok := bankNameCnMapping.dataMap[key]; ok {
		return nil
	}
	if config.DataDir == "" {
		return errors.New("存储地址未配置")
	}

	if err := ioutil.WriteFile(
		fmt.Sprintf("%s/%s", config.DataDir, bankNameCnFileName),
		[]byte(strings.Join([]string{key, name}, "=")), 0644); err != nil {
		logger.Error(err.Error())
		return errors.New("创建银行名称映射关系失败")
	}
	bankNameCnMapping.dataMap[key] = name
	return nil
}

func CreateCountryCnNameMapping(key, name string) error {
	if _, ok := countryCnMapping.dataMap[key]; ok {
		return nil
	}
	if config.DataDir == "" {
		return errors.New("存储地址未配置")
	}

	if err := ioutil.WriteFile(
		fmt.Sprintf("%s/%s", config.DataDir, bankNameCnFileName),
		[]byte(strings.Join([]string{key, name}, "=")), 0644); err != nil {
		logger.Error(err.Error())
		return errors.New("创建国家名称映射关系失败")
	}
	countryCnMapping.dataMap[key] = name
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
	if _, err = currentBinDatabase.ReadExact(uint32bin); err == nil {
		return nil
	}

	bindata.Id = time.Now().UnixNano()
	if err = currentBinDatabase.Save(uint32bin, bindata, approximate); err != nil {
		return err
	}
	return nil
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
		result      mod.BinData
		approximate bool
		ok          bool
	)
	if result, err = currentBinDatabase.ReadExact(uint32bin); err != nil || approximate {
		return nil, errors.New(fmt.Sprintf("bin %s not found", bin))
	}

	var bankNameCn, countryCn string
	if bankNameCn, ok = bankNameCnMapping.dataMap[result.BankName]; !ok {
		bankNameCn = result.BankName
	}
	if countryCn, ok = countryCnMapping.dataMap[result.Country]; !ok {
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

func readFromFile(event file.FileEvent) {
	filepath := event.Filepath
	filename := path.Base(filepath)
	if filename == bankNameCnFileName {
		bankNameCnMapping.reloading <- event
	} else if filename == countryCnFileName {
		countryCnMapping.reloading <- event
	} else {
		logger.Infof("忽略文件: %s\n", filepath)
	}
}

func refreshBankName(event file.FileEvent) {
	readMappingFile(bankNameCnMapping, event)
}

func refreshCountry(event file.FileEvent) {
	readMappingFile(countryCnMapping, event)
}

func readMappingFile(mpf mappingFile, e file.FileEvent) {
	var (
		f   *os.File
		err error
	)
	filepath := e.Filepath
	if f, err = os.Open(filepath); err != nil {
		logger.Errorf("read file error: %s\n", err.Error())
		return
	}
	if !e.FileCreated {
		if _, err := f.Seek(mpf.bytes, 0); err != nil {
			logger.Errorf("seek file %s error: %s\n", err.Error())
			return
		}
	}
	reader := bufio.NewReader(f)
	for {
		data, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		kv := strings.Split(string(data), "=")
		if _, ok := mpf.dataMap[kv[0]]; ok {
			continue
		}
		mpf.dataMap[kv[0]] = kv[1]
	}

	if fileInfo, err := f.Stat(); err == nil {
		mpf.bytes += fileInfo.Size()
	}
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
	config.DataDir = path.Dir(dir)
	go registerFileHander()
	prepare(dir)
	beginWatching(dir)
}

func registerFileHander() {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(error); ok {
				logger.Errorf("panic: %s\n", v.Error())
			}
			logger.Errorf("file handler error: %s\n", string(debug.Stack()))
		}
	}()
	for {
		select {
		case event := <-bankNameCnMapping.reloading:
			bankNameCnMapping.handler(event)
		case event := <-countryCnMapping.reloading:
			countryCnMapping.handler(event)
		}
	}
}

func prepare(dir string) {
	var (
		filepaths []string
		err       error
	)
	if filepaths, err = file.SearchDir(dir, func(filepath string) bool {
		filename := path.Base(filepath)
		return bankNameCnFileName == filename || countryCnFileName == filename
	}); err != nil {
		logger.Fatal("prepare data failed, error: %s", err.Error())
	}

	for _, filepath := range filepaths {
		readFromFile(file.FileEvent{Filepath: filepath, FileCreated: true})
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
					e := file.FileEvent{Filepath: event.Name, FileCreated: false}
					readFromFile(e)
					for _, l := range fileEventListenerList {
						l(e)
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					logger.Infof("file created %s:\n", event.Name)
					e := file.FileEvent{Filepath: event.Name, FileCreated: true}
					readFromFile(e)
					for _, l := range fileEventListenerList {
						l(e)
					}
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
