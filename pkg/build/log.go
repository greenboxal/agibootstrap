package build

type Log struct {
}

func (l *Log) Close() error {
	return nil
}

func (l *Log) Error(err error) {

}

func NewLog(outputFile string) (*Log, error) {
	return &Log{}, nil
}
