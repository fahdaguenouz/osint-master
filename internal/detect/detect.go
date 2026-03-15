package detect

type DetectedType int

const (
	TypeUnknown DetectedType = iota
	TypeFullName
	TypeIP
	TypeUsername
)

func Detect(input string) DetectedType {
	switch {
	case IsIPv4(input):
		return TypeIP
	case IsFullName(input):
		return TypeFullName
	case IsUsername(input):
		return TypeUsername
	default:
		return TypeUnknown
	}
}
