package config

type HTTPServer struct {
	Host string
	Port string
}

type Config struct {
	HTTPServer HTTPServer
}

func Load() *Config {
	s := HTTPServer{Host: "0.0.0.0", Port: "8080"}
	c := &Config{HTTPServer: s}
	return c
}
