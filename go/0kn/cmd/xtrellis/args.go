package main

type ArgsCommon struct {
	////////////////////////////////////
	// files
	////////////////////////////////////
	ServerFile  string `default:"servers.json"`
	GroupFile   string `default:"groups.json"`
	ClientFile  string `default:"clients.json"`
	KeyFile     string `default:"keys.json"`
	MessageFile string `default:"messages.json"`
	OutFile     string `default:"res.json"`
}

type ArgsCoordinator struct {
	ArgsCommon

	GatewayAddrIn  string `default:"localhost:9000" help:"gateway proxy address for incoming mix-net messages"`
	GatewayAddrOut string `default:"localhost:9900" help:"gateway proxy address for outgoing mix-net messages"`
	GatewayEnable  bool   `default:"False" help:"enable client message gateway"`
	RoundInterval  int    `default:"0" help:"delay (in ms) between mix-net lightning rounds"`
	RunExperiment  bool   `default:"False" help:"run coordinator experiment"`

	F           float64 `default:"0"`
	RunType     int     `default:"1"`
	NumUsers    int     `default:"100" help:"also NumMessages"`
	NumServers  int     `default:"10"`
	MessageSize int     `default:"1024"`

	NumGroups int `default:"3"`
	GroupSize int `default:"3"`
	BinSize   int `default:"0"`
	LimitSize int `default:"0"`
	NumLayers int `default:"10"`
	Overflow  int `default:"32"`

	NumClientServers int  `default:"0"`
	SkipPathGen      bool `default:"False"`
	NoCheck          bool `default:"False"`
	LoadMessages     bool `default:"False"`
	StartIdx         int  `default:"0"`
	Interval         int  `default:"0"`
}

type ArgsServer struct {
	ArgsCommon

	Addr string `default:"localhost:8000"`
}

type ArgsClient struct {
	ArgsCommon

	Addr string `default:"localhost:8000"`
}

type Args struct {
	// commands
	Coordinator *ArgsCoordinator `arg:"subcommand:coordinator"`
	Server      *ArgsServer      `arg:"subcommand:server"`
	Client      *ArgsClient      `arg:"subcommand:client"`

	// global options
	Debug       bool `default:"False" help:"enable debug log output"`
	DebugCaller bool `default:"False" help:"with debug enabled, print calling function's info"`
}
