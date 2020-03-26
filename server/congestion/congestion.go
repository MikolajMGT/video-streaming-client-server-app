package congestion

const (
	NoCongestion       = 0
	LowCongestion      = 1
	MediumCongestion   = 2
	HighCongestion     = 3
	VeryHighCongestion = 4
)

func ResolveCongestionLevel(fractionLost float64) int {
	if fractionLost >= 0 && fractionLost <= 0.01 {
		return NoCongestion
	} else if fractionLost > 0.01 && fractionLost <= 0.25 {
		return LowCongestion
	} else if fractionLost > 0.25 && fractionLost <= 0.5 {
		return MediumCongestion
	} else if fractionLost > 0.5 && fractionLost <= 0.75 {
		return HighCongestion
	} else {
		return VeryHighCongestion
	}
}
