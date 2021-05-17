package main

import (
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/thanos-io/thanos/pkg/block/metadata"
	"github.com/thanos-io/thanos/pkg/compact/downsample"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "")
	app.Version("v0.1")
	app.HelpFlag.Short('h')

	defaultDBPath := "data/"
	// 1m.
	defaultResolution := "60000"
	defaultDownsampleDir := "/tmp/thanos-downsample"

	dbPath := app.Arg("db path", "Database path (default is "+defaultDBPath+").").Default(defaultDBPath).String()
	blockID := app.Arg("block id", "Block to downsample").Required().String()
	resolution := app.Flag("res", "Downsample resolution (default is 1m)").Default(defaultResolution).Int64()
	downsampleDir := app.Flag("dir", "Directory for downsampling").Default(defaultDownsampleDir).String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	logger := log.NewLogfmtLogger(os.Stdout)
	blockDir := filepath.Join(*dbPath, *blockID)
	block, err := tsdb.OpenBlock(logger, blockDir, chunkenc.NewPool())
	if err != nil {
		level.Error(logger).Log("msg", "failed to open block", "id", *blockID, "error", err)
		os.Exit(1)
	}
	meta, err := metadata.ReadFromDir(blockDir)
	if err != nil {
		level.Error(logger).Log("msg", "failed to read meta", "id", *blockID, "error", err)
		os.Exit(1)
	}
	if _, err := downsample.Downsample(logger, meta, block, *downsampleDir, *resolution); err != nil {
		level.Error(logger).Log("msg", "failed to downsample", "id", *blockID, "error", err)
		os.Exit(1)
	}
}
