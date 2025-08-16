package ssh

type Server struct{}

type Config struct {
	Net  string `scfg:"net"`
	Addr string `scfg:"addr"`
	Key  string `scfg:"key"`
	Root string `scfg:"root"`
}
