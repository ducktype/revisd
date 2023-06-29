package main

import "fmt"
import "flag"
import "os"
import "os/signal"
import "os/user" //FORCE DYNAMIC LINKING IF BUILD COMMAND DOES NOT HAVE SPECIAL FLAGS: https://github.com/golang/go/issues/26492
import "net" //FORCE DYNAMIC LINKING IF BUILD COMMAND DOES NOT HAVE SPECIAL FLAGS: https://github.com/golang/go/issues/26492
//import "regexp"
import "runtime"
import "runtime/debug"
import "log"
import "log/syslog"
import "path/filepath"
import "syscall"
import "time"
//import "sync"
//import "debug/elf"
//import "bytes"
//import "errors"
//import tt "text/template"
//import ht "html/template"

/*---------------------------------------------------------------------*/

//official experimental golang libs, slices package will be included in golang 1.21+
//import "golang.org/x/exp/slices"
import "golang.org/x/sys/unix"

/*---------------------------------------------------------------------*/

//some useful third-party modules:

//https://kendru.github.io/go/2021/10/26/sorting-a-dependency-graph-in-go/
//import "github.com/gonum/gonum" //topological sorting

//import "github.com/opcoder0/fanotify"
//https://github.com/s3rj1k/go-fanotify

import "github.com/davecgh/go-spew/spew"
//import cp "github.com/otiai10/copy"
//import _ "github.com/edwingeng/deque/v2"
//import "github.com/wk8/go-ordered-map"
//import _ "github.com/sourcegraph/conc"
//import _ "github.com/go-co-op/gocron"
//import _ "github.com/smallnest/chanx"

/*---------------------------------------------------------------------*/

//my own module local packages
import "revisord/util"


/*---------------------------------------------------------------------*/

type RevisorConf struct {
	exe_dir string
	conf_dir string
	ctrl_sock string
	args flag.FlagSet
}

func client(conf RevisorConf) {
	//var arg1 = 
	net.Dial("unix", conf.ctrl_sock)
}

func daemon(conf RevisorConf) {
	gg := util.NewGog()

	//quit channel


	//https://github.com/shoenig/nomad-pledge-driver/issues/2

	//set ourself as subreaper
	unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0);

	//handle signals
	var sigs_chan = make(chan os.Signal)
	signal.Notify(sigs_chan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCHLD)
	gg.Go(func() {
		time.Sleep(5 * time.Second)
	})

	//listen ctrl socket
	//l, _ := net.Listen("unix", conf.ctrl_sock)



/*
	//load conf
	fal, _ := fanotify.NewListener("/", false, fanotify.PermissionNone)
	var evt_types = fanotify.FileCreated | fanotify.FileModified | fanotify.FileDeleted | fanotify.FileMovedTo
	fal.AddWatch(conf.conf_dir, evt_types)
	go fal.Start()
	for evt_types := range fal.Events {
		fmt.Println(evt_types)
	}
	fal.Stop()

	*/


	/*
	event := Event{
		Fd:         fd,
		Path:       pathName,
		FileName:   fileName,
		EventTypes: EventType(mask),
		Pid:        int(metadata.Pid),
	}
	l.Events <- event
	*/

	gg.Wait(true)

}



func main() {
	
	var conf RevisorConf

	conf.exe_dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.conf_dir = filepath.Join(conf.exe_dir,".revisord")
	conf.ctrl_sock = filepath.Join(conf.conf_dir,"ctrl.sock")

	//configure cli arguments
	var flagset = flag.NewFlagSet("revisord", flag.ExitOnError)
	var f_daemon = flagset.Bool("d", true, "run as daemon")
	flagset.Parse(os.Args[1:])

	if !*f_daemon {
		client(conf)
	} else {
		daemon(conf)
	}

	

	syslogger, _ := syslog.New(syslog.LOG_INFO, "syslog_example")
  log.SetOutput(syslogger)
  log.Println("Log entry")


	gcpp := debug.SetGCPercent(-1)

	
	//set ourself as subreaper
	unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0);

	//https://github.com/golang/go/issues/44312
	//https://bugzilla.kernel.org/show_bug.cgi?id=211919
	//However version 2 and version 3 capabilities are 64 bit and do not fit into a single unix.CapUserData.
	//Instead of using a struct with wider fields however linux capget syscall uses the same struct but writes into 2 instances of the struct.
	//It writes lower bits of the capabilities into the first struct and the higher bits into the second one.
	runtime.LockOSThread() // exclusively bind the goroutine to a single os thread.
	hdr := unix.CapUserHeader{Version: unix.LINUX_CAPABILITY_VERSION_3}
	var data [2]unix.CapUserData
	data[0].Effective |= 1<<unix.CAP_NET_BIND_SERVICE
	//data[0].Effective |= 1<<unix.CAP_SYS_ADMIN
	//data[1].Effective |= 1<<(unix.CAP_CHECKPOINT_RESTORE - 32)
	unix.Capset(&hdr, &data[0])
	spew.Dump(hdr)
	spew.Dump(data)

	unix.Setuid(12)
	unix.Setgid(13)
	
	

	//PIDFD CLONE3 SETCGROUP
	//https://github.com/golang/go/issues/51246
	

	u, _ := user.Lookup("mysql")
	g, _ := user.LookupGroup("mysql")
	spew.Dump(u)
	spew.Dump(g)


	runtime.UnlockOSThread()
	debug.SetGCPercent(gcpp)

	/*
	//print arguments
	fmt.Printf("cfg_path: %q\n", *cfg_path)
	
	//read configuration file
	b, err := os.ReadFile(*cfg_path)
  if err != nil {
      fmt.Print(err)
  }
  var cfg_data = string(b)
	//fmt.Println(cfg_data)

	//configuration variables
	var cfg = conf_parse(cfg_data,"path")
	//fmt.Printf("cfg: %q\n", cfg)
	spew.Dump(cfg)
	*/


  fmt.Println("end")
}
