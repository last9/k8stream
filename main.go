package main

import (
	fmt "fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var (
	configFile = kingpin.Flag("config", "Config File to Parse").Required().File()
)

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

	ch := NewBatch(
		conf.UID, conf.BatchSize, conf.BatchInterval, sink, mcache,
	)

	h := &Handler{kc, ch, mcache}

	stopCh := make(chan struct{})
	informer.AddEventHandler(h)
	go informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	<-sigCh
	close(stopCh)
}
