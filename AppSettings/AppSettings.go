package as

type AppSettings struct {
	Debug                AppDebugSettings
	LockThread           bool
	Window               AppWindowSettings
	PrintJobsChannelSize int
	ImagesToLoad         []string
}
