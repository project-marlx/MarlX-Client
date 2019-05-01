package config

type ClientConfig struct {
	Store_dir string `bson:"store_dir"`
	MTU       uint64 `bson:"MTU"`
	Token     string `bson:"token"`
}

type ClientHandle struct {
	Quit    bool       `bson:"quit"`
	Channel chan error `bson:"channel"`
}
