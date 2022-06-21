package level

type Level interface {
	Code() int
	String() string
}

type level struct {
	code int
	name string
}

func (l level) Code() int {
	return l.code
}

func (l level) String() string {
	return l.name
}

var _ Level = &level{}

var (
	LevelFatal   Level = &level{code: 0, name: "fatal"}
	LevelError   Level = &level{code: 1, name: "error"}
	LevelWarning Level = &level{code: 2, name: "warning"}
	LevelInfo    Level = &level{code: 3, name: "info"}
	LevelNotice  Level = &level{code: 4, name: "notice"}
	LevelDebug   Level = &level{code: 5, name: "debug"}
	LevelTrace   Level = &level{code: 6, name: "trace"}
)

var levels = []Level{LevelFatal, LevelError, LevelWarning, LevelInfo, LevelNotice, LevelDebug, LevelTrace}

func GetLevelByName(name string) Level {
	for _, l := range levels {
		if l.String() == name {
			return l
		}
	}
	return LevelInfo
}

func GetLevelByCode(code int) Level {
	for _, l := range levels {
		if l.Code() == code {
			return l
		}
	}
	return LevelInfo
}
