package main

import (
	fmt "fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var (
	configFile = kingpin.Flag("config", "Config File to Parse").Required().File()
	wg         = &sync.WaitGroup{}
)

const VERSION = "0.0.1"

func main() {
	kingpin.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cData, err := readConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	conf := &L9K8streamConfig{}
	if err := loadConfig(cData, conf); err != nil {
		log.Fatal(err)
	}

	kc, err := newK8sClient(conf.KubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	factory := informers.NewSharedInformerFactory(kc.Clientset, time.Duration(60)*time.Second)
	informer := factory.Core().V1().Events().Informer()

	mcache, err := cacheClient()
	if err != nil {
		log.Fatal(err)
	}

	sink, err := getFlusher(conf, cData)
	if err != nil {
		log.Fatal(err)
	}

	flusherStopCh := make(chan struct{})
	ch := NewBatch(conf.UID, conf.BatchSize, conf.BatchInterval, sink, mcache, flusherStopCh, wg)

	h := &Handler{kc, ch, mcache}

	stopCh := make(chan struct{})
	informer.AddEventHandler(h)
	go informer.Run(stopCh)

	// TODO: currently, we are passing 2 channels and wg just for graceful shutdown. this needs to be changed.
	// TODO: the heartbeat function should not be concerned with the chan and wg values
	if err := StartHeartbeat(
		conf.UID,
		conf.HeartbeatHook,
		conf.HeartbeatInterval,
		conf.HeartbeatTimeout,
		stopCh,
		flusherStopCh,
		wg,
		func(stopCh chan struct{}, flusherStopCh chan struct{}, wg *sync.WaitGroup) {
			fmt.Errorf("upgrade required for k8stream version %s, halting event pull", VERSION)
			close(stopCh) // stop receiving k8s events
			wg.Add(1)
			close(flusherStopCh) // waits till currents events are flushed to sink, then stops consuming.
			wg.Wait()
			os.Exit(1)
		}); err != nil {
		log.Fatal(err)
	}

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	<-sigCh
	close(stopCh)
}
