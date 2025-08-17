package lmtp

type Config struct {
	Socket       string `scfg:"socket"`
	Domain       string `scfg:"domain"`
	MaxSize      int64  `scfg:"max_size"`
	WriteTimeout uint32 `scfg:"write_timeout"`
	ReadTimeout  uint32 `scfg:"read_timeout"`
}
