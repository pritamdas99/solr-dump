package solr_dump

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	dbc "kubedb.dev/db-client-go/solr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	scm = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientSetScheme.AddToScheme(scm))
	utilruntime.Must(api.AddToScheme(scm))
}

type SolrDump struct {
	action     string
	slClient   dbc.SLClient
	location   string
	repository string
}

func NewSolrDump(action string, dbname string, namespace string, location string, repository string) (*SolrDump, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Failed ti get config %s", config)
		return nil, err
	}
	kc, err := client.New(config, client.Options{
		Scheme: scm,
		Mapper: nil,
	})
	if err != nil {
		fmt.Println("failed to get client")
		return nil, err
	}
	db := &api.Solr{}
	err = kc.Get(context.TODO(), types.NamespacedName{
		Name:      dbname,
		Namespace: namespace,
	}, db)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	slClient, err := dbc.NewKubeDBClientBuilder(kc, db).WithContext(ctx).WithLog(klog.Background()).GetSolrClient()
	if err != nil {
		return nil, err
	}
	return &SolrDump{
		action,
		slClient,
		location,
		repository,
	}, nil
}

func (dumper *SolrDump) Execute() {
	if dumper.action == "backup" {
		err := dumper.backup()
		if err != nil {
			klog.Error(err)
		}
	}
}

func (dumper *SolrDump) backup() error {
	resp, err := dumper.slClient.ListCollection()
	if err != nil {
		return err
	}

	responseBody, err := dumper.slClient.DecodeResponse(resp)
	if err != nil {
		return err
	}

	_, err = dumper.slClient.GetResponseStatus(responseBody)
	if err != nil {
		klog.Error(fmt.Sprintf("status is non zero while listing collection"))
		return err
	}

	collectionList, err := dumper.slClient.GetCollectionList(responseBody)
	if err != nil {
		return err
	}

	for _, collection := range collectionList {
		if collection == "kubedb-system" {
			continue
		}
		fmt.Printf("backup collection %s", collection)
		resp, err := dumper.slClient.BackupCollection(context.TODO(), collection, fmt.Sprintf("%s-backup", collection), dumper.location, dumper.repository)
		if err != nil {
			klog.Error(fmt.Sprintf("Failed to backup collection %s", collection))
			return err
		}
		responseBody, err := dumper.slClient.DecodeResponse(resp)
		klog.Infof(fmt.Sprintf("responsebody %v", responseBody))
		if err != nil {
			klog.Error(fmt.Sprintf("Failed to decode backup response body for collection %s", collection))
			return err
		}
		_, err = dumper.slClient.GetResponseStatus(responseBody)
		if err != nil {
			klog.Error(fmt.Sprintf("status is non zero while listing collection"))
			return err
		}
	}
	return nil
}
