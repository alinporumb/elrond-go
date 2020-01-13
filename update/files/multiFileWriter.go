package files

import (
	"bufio"
	"os"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var log = logger.GetOrCreate("update/files")

type multiFileWriter struct {
	exportFolder string

	files       map[string]*os.File
	dataWriters map[string]*bufio.Writer
	exportStore storage.Storer
}

type ArgsNewMultiFileWriter struct {
	ExportFolder string
	ExportStore  storage.Storer
}

func NewMultiFileWriter(args ArgsNewMultiFileWriter) (*multiFileWriter, error) {
	if check.IfNil(args.ExportStore) {
		return nil, epochStart.ErrNilStorage
	}
	if len(args.ExportFolder) == 2 {
		return nil, update.ErrInvalidFolderName
	}

	m := &multiFileWriter{
		exportFolder: args.ExportFolder,
		files:        make(map[string]*os.File),
		dataWriters:  make(map[string]*bufio.Writer),
		exportStore:  args.ExportStore,
	}

	return m, nil
}

func (m *multiFileWriter) NewFile(name string) error {
	if _, ok := m.files[name]; ok {
		return nil
	}

	file, err := os.OpenFile(m.exportFolder+"/test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Debug("unable to open file")
		return err
	}

	dataWriter := bufio.NewWriter(file)
	m.files[name] = file
	m.dataWriters[name] = dataWriter

	return nil
}

func (m *multiFileWriter) Write(fileName string, key string, value []byte) error {
	dataWriter, ok := m.dataWriters[fileName]
	if !ok {
		err := m.NewFile(fileName)
		if err != nil {
			return nil
		}

		dataWriter, ok = m.dataWriters[fileName]
		if !ok {
			return update.ErrNilDataWriter
		}
	}

	_, err := dataWriter.WriteString(key + "\n")
	if err != nil {
		return err
	}

	err = m.exportStore.Put([]byte(key), value)
	if err != nil {
		return err
	}

	return nil
}

func (m *multiFileWriter) Finish() {
	for fileName, dataWriter := range m.dataWriters {
		err := dataWriter.Flush()
		if err != nil {
			log.Warn("could not flush data for ", "fileName", fileName, "error", err)
		}
	}

	for fileName, file := range m.files {
		err := file.Close()
		if err != nil {
			log.Warn("could not close file ", "fileName", fileName, "error", err)
		}
	}
}

func (m *multiFileWriter) IsInterfaceNil() bool {
	return m == nil
}