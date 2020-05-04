package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/last9/k8stream/io"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const VERSION = "0.0.3"

var (
	configFile = kingpin.Flag("config", "Config File to Parse").Required().File()
)

func getFlusher(conf *L9K8streamConfig) (io.Flusher, error) {
	return io.GetFlusher(&conf.Config)
}

func main() {
	kingpin.Version(VERSION)
	kingpin.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cData, err := io.ReadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	conf := &L9K8streamConfig{}
	if err := io.LoadConfig(cData, conf); err != nil {
		log.Fatal(err)
	}

	if err := io.StartHeartbeat(
		VERSION,
		conf.UID, conf.HeartbeatHook,
		conf.HeartbeatInterval, conf.HeartbeatTimeout,
	); err != nil {
		log.Fatal(err)
	}

	conf.Raw = cData
	setDefaults(conf)

	// Create a k8s client
	kc, err := newK8sClient(conf.KubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create a LRU Cache
	mcache, err := newCache()
	if err != nil {
		log.Fatal(err)
	}

	// Get Flusher instance from IO
	f, err := getFlusher(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Start a batcher, returns a channel.
	ch := startIngester(f, conf, mcache)
	h := &Handler{kc, ch, mcache, conf}

	stopCh := make(chan struct{})
	factory := informers.NewSharedInformerFactory(
		kc.Clientset,
		time.Duration(conf.ResyncInterval)*time.Second,
	)

	// Service Informer to capture service events, since they dont show up
	// in the defaults events interface.
	svcInformer := factory.Core().V1().Services().Informer()
	svcInformer.AddEventHandler(h)
	go svcInformer.Run(stopCh)

	informer := factory.Core().V1().Events().Informer()
	informer.AddEventHandler(h)
	go informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	os.Exit(trapSignal(stopCh))
}

func trapSignal(stopCh chan<- struct{}) int {
	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt, syscall.SIGQUIT)

	s := <-sigCh
	close(stopCh)

	if s == syscall.SIGQUIT {
		time.Sleep(300 * time.Millisecond)
		return 1
	}

	return 0
}
