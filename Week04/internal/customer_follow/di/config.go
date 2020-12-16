package di

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
}

type Server struct {
	Addr string `json:"addr"`
}

type Database struct {
	DSN                string `json:"dsn"`
	MaxIdleConns       *int   `json:"max_idle_conns"`
	MaxOpenConns       *int   `json:"max_open_conns"`
	ConnMaxIdleTimeSec *int64 `json:"conn_max_idle_time_sec"`
}
