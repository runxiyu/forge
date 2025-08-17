package ssh

type Config struct {
	Net             string `scfg:"net"`
	Addr            string `scfg:"addr"`
	Key             string `scfg:"key"`
	Root            string `scfg:"root"`
	ShutdownTimeout uint32 `scfg:"shutdown_timeout"`
}
