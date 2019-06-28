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
	bankNameCnMapping      = mappingFile{reloading: make(chan file.FileEvent), dataMap: make(map[string]string, 3000)}
	countryCnMapping       = mappingFile{reloading: make(chan file.FileEvent), dataMap: make(map[string]string, 300)}
	NullBinData            mod.BinData
	setBinDatabaseModeOnce sync.Once
)

type fileEventListener func(file.FileEvent)

var fileEventListenerList []fileEventListener

type BinDataConfig struct {
	DataDir string
}

type mappingFile struct {
	reloading chan file.FileEvent
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
			logger.Fatal("refresh bin data error: %s", err)
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

	filepath := fmt.Sprintf("%s/%s", config.DataDir, bankNameCnFileName)
	var (
		file *os.File
		err  error
	)

	if file, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		return err
	}

	defer func() {
		file.Close()
	}()

	if _, err = file.WriteString(strings.Join([]string{key, name}, "=")); err != nil {
		return err
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

	filepath := fmt.Sprintf("%s/%s", config.DataDir, bankNameCnFileName)

	var (
		file *os.File
		err  error
	)

	if file, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		return err
	}

	defer func() {
		file.Close()
	}()

	if _, err = file.WriteString(strings.Join([]string{key, name}, "=")); err != nil {
		return err
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
		logger.Infof("忽略文件: %s", filepath)
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
		logger.Errorf("read file error: %s", err)
		return
	}
	if !e.FileCreated {
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			logger.Errorf("seek file %s error: %s", err)
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
}

func WatchBinDataDir(dir string) {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(error); ok {
				logger.Errorf("panic: %s\n", v.Error())
			}
			logger.Errorf("watching bin data directory error: %s", string(debug.Stack()))
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
			logger.Errorf("file handler error: %s", string(debug.Stack()))
		}
	}()
	for {
		select {
		case event := <-bankNameCnMapping.reloading:
			refreshBankName(event)
		case event := <-countryCnMapping.reloading:
			refreshCountry(event)
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
		logger.Errorf("prepare data failed, error: %s", err)
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
					logger.Infof("file modified %s:", event.Name)
					e := file.FileEvent{Filepath: event.Name, FileCreated: false}
					readFromFile(e)
					for _, l := range fileEventListenerList {
						l(e)
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					logger.Infof("file created %s:", event.Name)
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
				logger.Error("watch %s error:%s", dir, err)
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
