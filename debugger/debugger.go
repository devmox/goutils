package debugger

import (
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// Debugger ...
type Debugger struct {
	trace     *log.Logger
	trace2    *log.Logger
	info      *log.Logger
	war       *log.Logger
	err       *log.Logger
	timeStart map[string]time.Time
	timeEnd   map[string]time.Time
	mux       sync.Mutex
	debug     int64
}

var debugger *Debugger

// GetDebugger retourne l'instance (*Debugger)en cours d'utilisation accessible depuis tous les packages
func GetDebugger() *Debugger {
	return debugger
}

// InitDebugger initalise le logger.
func InitDebugger(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer, debug int64) {
	d := &Debugger{}

	d.trace = log.New(traceHandle, "[TRACE] ", log.Ldate|log.Ltime|log.Lshortfile)
	d.trace2 = log.New(infoHandle, "[TIME] ", log.Ldate|log.Ltime)
	d.info = log.New(infoHandle, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	d.war = log.New(warningHandle, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile)
	d.err = log.New(errorHandle, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	d.debug = debug

	d.timeStart = make(map[string]time.Time)
	d.timeEnd = make(map[string]time.Time)

	d.mux = sync.Mutex{}

	debugger = d
}

// Trace ...
func (d *Debugger) Trace(str string) {
	_, file, line, _ := runtime.Caller(1)
	d.trace.Println(".", file, ":", line, ". Trace message : ", str)
}

// Warning ...
func (d *Debugger) Warning(str string) {
	_, file, line, _ := runtime.Caller(1)
	d.war.Println(".", file, ":", line, ". Warning message : ", str)
}

// Info ...
func (d *Debugger) Info(str string) {
	_, file, line, _ := runtime.Caller(1)
	d.info.Println(".", file, ":", line, ". Info message : ", str)
}

// Error ...
func (d *Debugger) Error(err error) {
	_, file, line, _ := runtime.Caller(1)
	d.err.Println(".", file, ":", line, ". Error message : ", err)
}

// Fatalln affiche l'erreur puis stoppe le programme avec os.Exit(1).
func (d *Debugger) Fatalln(err error) {
	_, file, line, _ := runtime.Caller(1)
	d.err.Println(".", file, ":", line, ". Error message : ", err)
	// Ne pas modifier ce code, d'autres packages en déprend.
	os.Exit(1)
}

// Start ...
// IMPORTANT :
// Pour faire face au problème de concurrence dans le package "logger.Start()"" et "logger.End()",
// il faut utiliser "sync.Mutex", cela a pour effet de verrouiller l’accès à la carte durant l’écriture.
// Cette solution est temporaire, car en cas de fortes charges, les requêtes mettent du temps à répondre.
func (d *Debugger) Start(key string) {
	if d.debug == 1 {
		d.mux.Lock()
		defer d.mux.Unlock()

		d.timeStart[key] = time.Now()
	}
}

// End ...
func (d *Debugger) End(key string) {
	if d.debug == 1 {
		d.mux.Lock()
		defer d.mux.Unlock()

		d.timeEnd[key] = time.Now()

		d.trace2.Printf("[%s] %v\n", key, d.timeEnd[key].Sub(d.timeStart[key]))
	}
}

// EndGet renvoie le temps d'exécution.
// Exemple : x.Debugger.Start("GoBackup.Run")
// time := x.Debugger.EndGet("GoBackup.Run")
func (d *Debugger) EndGet(key string) time.Duration {
	if d.debug == 1 {
		d.mux.Lock()
		defer d.mux.Unlock()
		d.timeEnd[key] = time.Now()

		//return fmt.Sprintf("%.3fs. Native : %s", d.timeEnd[key].Sub(d.timeStart[key]).Seconds(), d.timeEnd[key].Sub(d.timeStart[key]))
		return d.timeEnd[key].Sub(d.timeStart[key])
	}

	return 0
}
