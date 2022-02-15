package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "FINANCIFY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =============================================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	if err := conf.Parse(os.Args[1:], "FINANCIFY", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("FINANCIFY", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}

			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("FINANCIFY", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}

			fmt.Println(version)
			return nil
		}

		return errors.Wrap(err, "parsing config")
	}

	expvar.NewString("build").Set(build)
	log.Printf("main: started: application initializing : version %q\n", build)
	defer log.Println("main: completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: config: \n%v\n", out)

	// =============================================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to default mux by importing the net/http/pprof packege.
	// /debug/vars - Added to default mux by importing the expvar packege.

	log.Println("main: initializing debugging support")

	go func() {
		log.Printf("main: debug listening %s", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: debug listener closed: %v", err)
		}
	}()

	select {}

	return nil
}
