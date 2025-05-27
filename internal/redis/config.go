// This package provides functions to manage the configuration of a Redis server
package redis

var (
	dir        string // Directory where the Redis server is running
	dbFileName string // Name of the Redis database file
)

func SetDirectory(d string) {
	dir = d
}

func SetDBFileName(name string) {
	dbFileName = name
}

func GetDirectory() string {
	return dir
}

func GetDBFileName() string {
	return dbFileName
}

func GetConfigFile() string {
	return dir + "/" + dbFileName
}
