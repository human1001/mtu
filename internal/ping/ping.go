package ping

// PingDF
// -1:太小, 1L 太大, 0: 出错, 此时error有值
// 当fastMode为true、int大于1、error等于空值时, 为直接返回得到MTU的值
func PingDF(l int, pingHost string, fastMode bool) (int, error) {
	return subPingDF(l, pingHost, fastMode)
}

func Wrap() string {
	return wrap()
}
