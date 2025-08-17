package irc

// Config contains IRC connection and identity settings for the bot.
// This should usually be a part of the primary config struct.
type Config struct {
	Net   string `scfg:"net"`
	Addr  string `scfg:"addr"`
	TLS   bool   `scfg:"tls"`
	SendQ uint   `scfg:"sendq"`
	Nick  string `scfg:"nick"`
	User  string `scfg:"user"`
	Gecos string `scfg:"gecos"`
}
